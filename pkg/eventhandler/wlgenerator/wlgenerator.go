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

package wlgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	membermanager "github.com/horizoncd/horizon/pkg/member"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	"gopkg.in/yaml.v3"

	"github.com/horizoncd/horizon/core/common"
	webhookctl "github.com/horizoncd/horizon/core/controller/webhook"
	"github.com/horizoncd/horizon/lib/q"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/event/models"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	webhookmanager "github.com/horizoncd/horizon/pkg/webhook/manager"
	webhookmodels "github.com/horizoncd/horizon/pkg/webhook/models"
)

const (
	WebhookSecretHeader      = "X-Horizon-Webhook-Secret"
	WebhookContentTypeHeader = "Content-Type"
	WebhookContentType       = "application/json;charset=utf-8"
)

// MessageContent will be marshaled as webhook request body
type MessageContent struct {
	ID          uint                  `json:"id,omitempty"`
	EventID     uint                  `json:"eventID,omitempty"`
	WebhookID   uint                  `json:"webhookID,omitempty"`
	Application *ApplicationInfo      `json:"application,omitempty"`
	Cluster     *ClusterInfo          `json:"cluster,omitempty"`
	Pipelinerun *PipelinerunInfo      `json:"pipelinerun,omitempty"`
	Member      *MemberInfo           `json:"member,omitempty"`
	EventType   string                `json:"eventType,omitempty"`
	User        *usermodels.UserBasic `json:"user,omitempty"`
	Extra       *string               `json:"extra,omitempty"`
}

type ResourceCommonInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name,omitempty"`
}

// ApplicationInfo contains basic info of application
type ApplicationInfo struct {
	ResourceCommonInfo
	Priority string `json:"priority,omitempty"`
}

// ClusterInfo contains basic info of cluster
type ClusterInfo struct {
	ResourceCommonInfo
	ApplicationName string `json:"applicationName,omitempty"`
	Env             string `json:"env,omitempty"`
}

// PipelinerunInfo contains basic info of pipelinerun
type PipelinerunInfo struct {
	ResourceCommonInfo
	ClusterID   uint   `json:"clusterID,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	Action      string `json:"action,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	GitRef      string `json:"gitRef,omitempty"`
	GitRefType  string `json:"gitRefType,omitempty"`
}

// MemberInfo contains basic info of member
type MemberInfo struct {
	ResourceCommonInfo
	ResourceID   uint
	ResourceType membermodels.ResourceType
	Role         string
	MemberNameID uint
}

// WebhookLogGenerator generates webhook logs by events
type WebhookLogGenerator struct {
	webhookMgr     webhookmanager.Manager
	eventMgr       eventmanager.Manager
	groupMgr       groupmanager.Manager
	applicationMgr applicationmanager.Manager
	clusterMgr     clustermanager.Manager
	prMgr          *prmanager.PRManager
	memberMgr      membermanager.Manager
	userMgr        usermanager.Manager
}

func NewWebhookLogGenerator(manager *managerparam.Manager) *WebhookLogGenerator {
	return &WebhookLogGenerator{
		webhookMgr:     manager.WebhookMgr,
		eventMgr:       manager.EventMgr,
		groupMgr:       manager.GroupMgr,
		applicationMgr: manager.ApplicationMgr,
		clusterMgr:     manager.ClusterMgr,
		prMgr:          manager.PRMgr,
		memberMgr:      manager.MemberMgr,
		userMgr:        manager.UserMgr,
	}
}

type messageDependency struct {
	webhook     *webhookmodels.Webhook
	event       *models.Event
	application *applicationmodels.Application
	cluster     *clustermodels.Cluster
	pipelinerun *prmodels.Pipelinerun
	member      *membermodels.Member
}

// listSystemResources lists root group(0) as system resource
func (w *WebhookLogGenerator) listSystemResources() map[string][]uint {
	return map[string][]uint{
		string(common.ResourceGroup): {0},
	}
}

// listAssociatedResourcesOfApp get application by id and list all the parent resources
func (w *WebhookLogGenerator) listAssociatedResourcesOfApp(ctx context.Context,
	id uint) (*applicationmodels.Application, map[string][]uint) {
	resources := w.listSystemResources()
	app, err := w.applicationMgr.GetByIDIncludeSoftDelete(ctx, id)
	if err != nil {
		log.Warningf(ctx, "application %d is not exist", id)
		return nil, resources
	}
	resources[common.ResourceApplication] = []uint{app.ID}

	group, err := w.groupMgr.GetByID(ctx, app.GroupID)
	if err != nil {
		log.Warningf(ctx, "application %d is not exist", id)
		return app, resources
	}
	groupIDs := groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs)
	resources[common.ResourceGroup] = append(resources[common.ResourceGroup], groupIDs...)
	return app, resources
}

// listAssociatedResourcesOfCluster get cluster by id and list all the parent resources
func (w *WebhookLogGenerator) listAssociatedResourcesOfCluster(ctx context.Context, id uint) (*clustermodels.Cluster,
	*applicationmodels.Application, map[string][]uint) {
	cluster, err := w.clusterMgr.GetByIDIncludeSoftDelete(ctx, id)
	if err != nil {
		log.Warningf(ctx, "cluster %d is not exist",
			id)
		return nil, nil, nil
	}
	app, resources := w.listAssociatedResourcesOfApp(ctx, cluster.ApplicationID)
	if resources == nil {
		resources = map[string][]uint{}
	}
	resources[common.ResourceCluster] = []uint{cluster.ID}
	return cluster, app, resources
}

// listAssociatedResourcesOfPipelinerun gets pipelinerun by id and list all the parent resources
func (w *WebhookLogGenerator) listAssociatedResourcesOfPipelinerun(ctx context.Context,
	id uint) (*prmodels.Pipelinerun, *clustermodels.Cluster, map[string][]uint) {
	pr, err := w.prMgr.PipelineRun.GetByID(ctx, id)
	if err != nil {
		log.Warningf(ctx, "pipelinerun %d is not exist",
			id)
		return nil, nil, nil
	}
	cluster, _, resources := w.listAssociatedResourcesOfCluster(ctx, pr.ClusterID)
	if resources == nil {
		resources = map[string][]uint{}
	}
	resources[common.ResourcePipelinerun] = []uint{pr.ID}
	return pr, cluster, resources
}

// listAssociatedResourcesOfMember gets member by id and list all the parent resources
func (w *WebhookLogGenerator) listAssociatedResourcesOfMember(ctx context.Context,
	id uint) (*membermodels.Member, map[string][]uint) {
	member, err := w.memberMgr.GetByIDIncludeSoftDelete(ctx, id)
	if err != nil {
		log.Warningf(ctx, "member %d is not exist", id)
		return nil, nil
	}
	var resources map[string][]uint
	switch member.ResourceType {
	case membermodels.TypeApplication:
		_, resources = w.listAssociatedResourcesOfApp(ctx, member.ResourceID)
	case membermodels.TypeApplicationCluster:
		_, _, resources = w.listAssociatedResourcesOfCluster(ctx, member.ResourceID)
	default:
		// TODO: support member event of groups and templates
		log.Warningf(ctx, "member event of resource type %s is unsupported yet", member.ResourceType)
	}
	return member, resources
}

// listAssociatedResources list all the associated resources of event to find all the webhooks
func (w *WebhookLogGenerator) listAssociatedResources(ctx context.Context,
	e *models.Event) (*messageDependency, map[string][]uint) {
	var (
		resources   map[string][]uint
		cluster     *clustermodels.Cluster
		application *applicationmodels.Application
		pr          *prmodels.Pipelinerun
		member      *membermodels.Member
		dep         = &messageDependency{}
	)

	switch e.ResourceType {
	case common.ResourceApplication:
		application, resources = w.listAssociatedResourcesOfApp(ctx, e.ResourceID)
		dep.application = application
	case common.ResourceCluster:
		cluster, application, resources = w.listAssociatedResourcesOfCluster(ctx, e.ResourceID)
		dep.application = application
		dep.cluster = cluster
	case common.ResourcePipelinerun:
		pr, cluster, resources = w.listAssociatedResourcesOfPipelinerun(ctx, e.ResourceID)
		dep.cluster = cluster
		dep.pipelinerun = pr
		log.Debugf(ctx, "dep: %+v", dep)
	case common.ResourceMember:
		member, resources = w.listAssociatedResourcesOfMember(ctx, e.ResourceID)
		dep.member = member
	default:
		log.Infof(ctx, "resource type %s is unsupported",
			e.ResourceType)
	}
	return dep, resources
}

// makeRequestHeaders assemble headers of webhook request
func (w *WebhookLogGenerator) makeRequestHeaders(secret string) (string, error) {
	header := http.Header{}
	header.Add(WebhookSecretHeader, secret)
	header.Add(WebhookContentTypeHeader, WebhookContentType)
	headerByte, err := yaml.Marshal(header)
	if err != nil {
		return "", err
	}
	return string(headerByte), nil
}

// makeRequestHeaders assemble body of webhook request
func (w *WebhookLogGenerator) makeRequestBody(ctx context.Context, dep *messageDependency) (string, error) {
	message := MessageContent{
		EventID:   dep.event.ID,
		WebhookID: dep.webhook.ID,
		EventType: dep.event.EventType,
		Extra:     dep.event.EventSummary.Extra,
	}

	if dep.event.CreatedBy != 0 {
		user, err := w.userMgr.GetUserByID(ctx, dep.event.CreatedBy)
		if err != nil {
			return "", err
		}
		message.User = usermodels.ToUser(user)
	}

	if dep.event.ResourceType == common.ResourceApplication &&
		dep.application != nil {
		message.Application = &ApplicationInfo{
			ResourceCommonInfo: ResourceCommonInfo{
				ID:   dep.application.ID,
				Name: dep.application.Name,
			},
			Priority: string(dep.application.Priority),
		}
	}

	if dep.event.ResourceType == common.ResourceCluster &&
		dep.cluster != nil {
		message.Cluster = &ClusterInfo{
			ResourceCommonInfo: ResourceCommonInfo{
				ID:   dep.cluster.ID,
				Name: dep.cluster.Name,
			},
			ApplicationName: dep.application.Name,
			Env:             dep.cluster.EnvironmentName,
		}
	}

	if dep.event.ResourceType == common.ResourcePipelinerun &&
		dep.pipelinerun != nil {
		message.Pipelinerun = &PipelinerunInfo{
			ResourceCommonInfo: ResourceCommonInfo{
				ID: dep.pipelinerun.ID,
			},
			ClusterID: dep.pipelinerun.ClusterID,
			ClusterName: func() string {
				if dep.cluster != nil {
					return dep.cluster.Name
				}
				return ""
			}(),
			Action:      dep.pipelinerun.Action,
			Title:       dep.pipelinerun.Title,
			Description: dep.pipelinerun.Description,
			GitRef:      dep.pipelinerun.GitRef,
			GitRefType:  dep.pipelinerun.GitRefType,
		}
	}

	if dep.event.ResourceType == common.ResourceMember &&
		dep.member != nil {
		message.Member = &MemberInfo{
			ResourceCommonInfo: ResourceCommonInfo{
				ID: dep.member.ID,
			},
			ResourceID:   dep.member.ResourceID,
			ResourceType: dep.member.ResourceType,
			Role:         dep.member.Role,
			MemberNameID: dep.member.MemberNameID,
		}
	}

	reqBody, err := json.Marshal(message)
	if err != nil {
		log.Errorf(ctx, fmt.Sprintf("failed to marshal message, error: %+v", err))
		return "", err
	}
	return string(reqBody), nil
}

// Process processes all the webhook logs that are in waiting status and send webhook requests
func (w *WebhookLogGenerator) Process(ctx context.Context, events []*models.Event,
	resume bool) error {
	var (
		webhookLogs        []*webhookmodels.WebhookLog
		conditionsToCreate = map[uint]map[uint]messageDependency{}
		conditionsToQuery  = map[uint][]uint{}
	)

	// TODO: simplify
	for _, event := range events {
		// 1. get associated resources according to event resource type
		dependency, resources := w.listAssociatedResources(ctx, event)
		if resources == nil {
			continue
		}

		// 2. list webhooks of all associated resources
		webhooks, _, err := w.webhookMgr.ListWebhookOfResources(ctx, resources, q.New(q.KeyWords{
			common.CreatedAt: event.CreatedAt,
			common.Enabled:   true,
		}))
		if err != nil {
			log.Errorf(ctx, "failed to list webhooks by condition %v, error: %s", resources, err.Error())
			continue
		}

		// 3. assemble webhook list of all events, prepare to create
		for _, webhook := range webhooks {
			// 3.1 if event does not match webhook trigger, skip
			ok, err := webhookctl.CheckIfEventMatch(webhook, event)
			if err != nil {
				log.Errorf(ctx, "failed to match triggers %s, error: %+v", webhook.Triggers, err)
				continue
			} else if !ok {
				continue
			}
			log.Debugf(ctx, "event %d matches webhook %s", event.ID, webhook.URL)
			// 3.2 add webhook to the list
			if _, ok := conditionsToCreate[event.ID]; !ok {
				conditionsToCreate[event.ID] = map[uint]messageDependency{}
			}
			conditionsToCreate[event.ID][webhook.ID] = messageDependency{
				webhook:     webhook,
				event:       event,
				application: dependency.application,
				cluster:     dependency.cluster,
				pipelinerun: dependency.pipelinerun,
				member:      dependency.member,
			}
			conditionsToQuery[event.ID] = append(conditionsToQuery[event.ID], webhook.ID)
		}
	}
	// 4. remove duplicate webhook log when resume
	if resume {
		if len(conditionsToQuery) == 0 {
			return nil
		}
		existedWebhookLogs, err := w.webhookMgr.ListWebhookLogsByMap(ctx, conditionsToQuery)
		if err != nil {
			log.Errorf(ctx, "failed to list webhook logs by map, error: %+v", err)
		} else {
			for _, wl := range existedWebhookLogs {
				delete(conditionsToCreate[wl.EventID], wl.WebhookID)
				if len(conditionsToCreate[wl.EventID]) == 0 {
					delete(conditionsToCreate, wl.EventID)
				}
			}
		}
	}

	// 5. assemble webhook logs to create
	if len(conditionsToCreate) == 0 {
		return nil
	}
	for _, dependencyMap := range conditionsToCreate {
		for _, dependency := range dependencyMap {
			headers, err := w.makeRequestHeaders(dependency.webhook.Secret)
			if err != nil {
				log.Errorf(ctx, fmt.Sprintf("failed to make headers, error: %+v", err))
				continue
			}

			body, err := w.makeRequestBody(ctx, &dependency)
			if err != nil {
				log.Errorf(ctx, fmt.Sprintf("failed to make headers, error: %+v", err))
				continue
			}

			webhookLogs = append(webhookLogs, &webhookmodels.WebhookLog{
				EventID:        dependency.event.ID,
				WebhookID:      dependency.webhook.ID,
				URL:            dependency.webhook.URL,
				RequestHeaders: headers,
				RequestData:    body,
				Status:         webhookmodels.StatusWaiting,
			})
		}
	}

	// 6. batch create webhook logs
	if _, err := w.webhookMgr.CreateWebhookLogs(ctx, webhookLogs); err != nil {
		log.Errorf(ctx, "failed to create webhooks, error: %s", err.Error())
		return err
	}
	return nil
}
