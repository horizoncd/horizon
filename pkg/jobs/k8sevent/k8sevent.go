// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sevent

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/config/k8sevent"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/regioninformers"
	"github.com/horizoncd/horizon/pkg/util/log"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	cacheMax                   = 160
	savingInterval             = 10 * time.Second
	kubernetesInstanceLabelKey = "app.kubernetes.io/instance"
)

var gvrEvent = schema.GroupVersionResource{
	Resource: "events",
	Version:  "v1",
	Group:    "",
}

type SuperVisor struct {
	filter    *gvkFilter
	informers *regioninformers.RegionInformers
	mgr       *managerparam.Manager
	db        *gorm.DB
	cacheMax  int
}

func New(config k8sevent.Config, informers *regioninformers.RegionInformers,
	mgr *managerparam.Manager, db *gorm.DB) *SuperVisor {
	v := &SuperVisor{
		filter:    newGVKFilter(config),
		informers: informers,
		mgr:       mgr,
		db:        db,
		cacheMax:  cacheMax,
	}

	return v
}

func (v *SuperVisor) Run(ctx context.Context) {
	v.informers.Register(regioninformers.Resource{GVR: gvrEvent, MakeHandler: v.newEventHandler})
}

type EventWithTime struct {
	*eventmodels.Event
	LastTimestamp time.Time
}

func (v *SuperVisor) newEventHandler(regionID uint, stopCh <-chan struct{}) (cache.ResourceEventHandler, error) {
	log.Debugf(context.Background(), "new event handler for region %d", regionID)
	ctx := context.Background()
	entity, err := v.mgr.RegionMgr.GetRegionByID(ctx, regionID)
	if err != nil {
		return nil, err
	}

	h := &eventHandler{
		SuperVisor: v,
		regionID:   regionID,
		region:     entity.Region,
		eventCh:    make(chan *corev1.Event),
		cacheMax:   v.cacheMax,
		stopCh:     stopCh,
	}

	go func() {
		eventCache := make([]*corev1.Event, 0, v.cacheMax)
		ticker := time.NewTicker(savingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if len(eventCache) == 0 {
					ticker.Reset(savingInterval)
					continue
				}
			case event := <-h.eventCh:
				eventCache = append(eventCache, event)
				if len(eventCache) < h.cacheMax {
					continue
				}
			}
			err := h.save(eventCache)
			if err != nil {
				log.Errorf(ctx, "failed to save event: %v", err)
			}
			eventCache = eventCache[:0]
			ticker.Reset(savingInterval)
		}
	}()

	return h, nil
}

type eventHandler struct {
	*SuperVisor
	regionID uint
	region   *models.Region
	cacheMax int

	stopCh <-chan struct{}

	eventCh chan *corev1.Event
}

func (e *eventHandler) save(cache []*corev1.Event) error {
	select {
	case <-e.stopCh:
		return nil
	default:
	}

	ctx := context.Background()
	eventsWithTime := make([]*EventWithTime, len(cache))

	reqIDs := sync.Map{}

	wg := sync.WaitGroup{}
	for i, event := range cache {
		wg.Add(1)
		go func(index int, event *corev1.Event, dst []*EventWithTime) {
			defer wg.Done()
			horizonEvent, err := e.mapEvent(event)
			if err != nil {
				log.Errorf(ctx, "failed to map event: %v", err)
				return
			}

			// skip same reqID
			if _, ok := reqIDs.Load(horizonEvent.ReqID); ok {
				return
			}

			reqIDs.Store(horizonEvent.ReqID, struct{}{})

			sameEvents, _ := e.mgr.EventManager.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.ReqID: horizonEvent.ReqID}})
			// skip same event
			for _, sameEvent := range sameEvents {
				if sameEvent.Extra != nil && horizonEvent.Extra != nil &&
					*sameEvent.Extra == *horizonEvent.Extra {
					return
				}
			}
			dst[index] = &EventWithTime{horizonEvent, event.LastTimestamp.Time}
		}(i, event, eventsWithTime)
	}
	wg.Wait()

	// skip nil in eventsWithTime
	for i := len(eventsWithTime) - 1; i >= 0; i-- {
		if eventsWithTime[i] == nil {
			eventsWithTime = append(eventsWithTime[:i], eventsWithTime[i+1:]...)
		}
	}

	if len(eventsWithTime) == 0 {
		return nil
	}

	sort.Slice(eventsWithTime, func(i, j int) bool {
		return eventsWithTime[i].LastTimestamp.Before(eventsWithTime[j].LastTimestamp)
	})

	tx := e.db.WithContext(ctx).Begin()
	txEventManager := eventmanager.New(tx)

	events := make([]*eventmodels.Event, 0, len(eventsWithTime))
	for _, event := range eventsWithTime {
		events = append(events, event.Event)
	}

	_, err := txEventManager.CreateEvent(ctx, events...)
	if err != nil {
		log.Errorf(ctx, "failed to save regionEvents: %v", err)
		tx.Rollback()
		return nil
	}

	err = tx.Commit().Error
	if err != nil {
		log.Errorf(ctx, "failed to commit: %v", err)
		tx.Rollback()
		return nil
	}
	return nil
}

func (*eventHandler) compactEvent(event *corev1.Event) map[string]interface{} {
	m := make(map[string]interface{})

	m["message"] = event.Message
	m["reason"] = event.Reason
	m["type"] = event.Type
	m["lastTimestamp"] = event.LastTimestamp
	m["name"] = event.Name

	obj := make(map[string]interface{})
	obj["kind"] = event.InvolvedObject.Kind
	obj["name"] = event.InvolvedObject.Name
	obj["namespace"] = event.InvolvedObject.Namespace
	obj["apiVersion"] = event.InvolvedObject.APIVersion
	m["involvedObject"] = obj
	return m
}

var GVKApplication = schema.GroupVersionKind{
	Group:   "argoproj.io",
	Version: "v1alpha1",
	Kind:    "Application",
}

func (e *eventHandler) mapEvent(event *corev1.Event) (*eventmodels.Event, error) {
	if event == nil {
		return nil, nil
	}
	extra, err := json.Marshal(e.compactEvent(event))
	if err != nil {
		log.Errorf(context.Background(), "failed to marshal event: %v", err)
		return nil, err
	}
	extraStr := string(extra)
	horizonEvent := &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterKubernetesEvent,
			Extra:        &extraStr,
		},
		ReqID: string(event.UID),
	}

	involvedObj := event.InvolvedObject

	var (
		obj  runtime.Object
		ns   = involvedObj.Namespace
		name = involvedObj.Name
		gvk  = involvedObj.GroupVersionKind()
	)

	if gvk == GVKApplication {
		cluster, err := e.mgr.ClusterMgr.GetByName(context.Background(), name)
		if err != nil {
			return nil, err
		}
		horizonEvent.ResourceID = cluster.ID
		return horizonEvent, nil
	}

	for {
		gvr, err := e.informers.GVK2GVR(e.regionID, gvk)
		if err != nil {
			return nil, err
		}
		err = e.informers.GetDynamicInformer(e.regionID, gvr, func(informer informers.GenericInformer) error {
			obj, err = informer.Lister().ByNamespace(ns).Get(name)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		un, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("failed to convert object to unstructured")
		}
		labels := un.GetLabels()
		if labels != nil && labels[kubernetesInstanceLabelKey] != "" {
			instanceName := labels[kubernetesInstanceLabelKey]
			cluster, err := e.mgr.ClusterMgr.GetByName(context.Background(), instanceName)
			if err != nil {
				return nil, err
			}
			horizonEvent.ResourceID = cluster.ID
			return horizonEvent, nil
		}
		ownerReferences := un.GetOwnerReferences()
		if len(ownerReferences) != 0 {
			ownerReference := ownerReferences[0]
			gvk = schema.FromAPIVersionAndKind(ownerReference.APIVersion, ownerReference.Kind)
			name = ownerReference.Name
		} else {
			return horizonEvent, nil
		}
	}
}

func (e *eventHandler) addEventToCache(un *unstructured.Unstructured) {
	event := &corev1.Event{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, event)
	if err != nil {
		log.Errorf(context.Background(), "failed to convert unstructured to event: %v", err)
		return
	}

	if !e.filter.has(event) {
		return
	}

	e.eventCh <- event
}

func (e *eventHandler) OnAdd(obj interface{}) {
	log.Debugf(context.Background(), "%p: event added\n", e)
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		log.Errorf(context.Background(), "failed to convert object to unstructured")
		return
	}
	e.addEventToCache(un)
}

func (e *eventHandler) OnUpdate(_, newObj interface{}) {
	log.Debugf(context.Background(), "%v: event updated", e)
	un, ok := newObj.(*unstructured.Unstructured)
	if !ok {
		log.Errorf(context.Background(), "failed to convert object to unstructured")
		return
	}
	e.addEventToCache(un)
}

func (e *eventHandler) OnDelete(_ interface{}) {
	// no need to handle delete event
}

type reasonIndex map[string][]*regexp.Regexp

type gvkIndex map[schema.GroupVersionKind]reasonIndex

type gvkFilter struct {
	index gvkIndex
}

func newGVKFilter(config k8sevent.Config) *gvkFilter {
	index := make(gvkIndex)

	for _, rule := range config.Rules {
		gvk := rule.GroupVersionKind
		if _, ok := index[gvk]; !ok {
			index[gvk] = make(reasonIndex)
		}
		for _, reason := range rule.Reasons {
			if _, ok := index[gvk][reason.Reason]; !ok {
				index[gvk][reason.Reason] = nil
			}
			for _, message := range reason.Messages {
				if message == "" {
					continue
				}
				pattern := regexp.MustCompile(message)
				index[gvk][reason.Reason] = append(index[gvk][reason.Reason], pattern)
			}
		}
	}

	return &gvkFilter{index}
}

func (f *gvkFilter) has(event *corev1.Event) bool {
	gvk := schema.FromAPIVersionAndKind(event.InvolvedObject.APIVersion, event.InvolvedObject.Kind)
	if reasonIndex, ok := f.index[gvk]; ok {
		if len(reasonIndex) == 0 {
			return true
		}
		if patterns, ok := reasonIndex[event.Reason]; ok {
			if len(patterns) == 0 {
				return true
			}
			for _, pattern := range patterns {
				if pattern.MatchString(event.Message) {
					return true
				}
			}
		}
	}
	log.Debugf(context.Background(),
		"event (%s) has been skipped: gvk = %s, uid = %s", event.Message, gvk.String(), event.UID)
	return false
}
