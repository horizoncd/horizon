package wlgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/horizoncd/horizon/lib/q"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/core/common"
	webhookctl "github.com/horizoncd/horizon/pkg/core/controller/webhook"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/event/models"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
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
	EventType   string                `json:"eventType,omitempty"`
	User        *usermodels.UserBasic `json:"user,omitempty"`
	Extra       *string               `json:"extra,omitempty"`
}

type ResourceCommonInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
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

// WebhookLogGenerator generates webhook logs by events
type WebhookLogGenerator struct {
	webhookMgr     webhookmanager.Manager
	eventMgr       eventmanager.Manager
	groupMgr       groupmanager.Manager
	applicationMgr applicationmanager.Manager
	clusterMgr     clustermanager.Manager
	userMgr        usermanager.Manager
}

func NewWebhookLogGenerator(manager *managerparam.Manager) *WebhookLogGenerator {
	return &WebhookLogGenerator{
		webhookMgr:     manager.WebhookManager,
		eventMgr:       manager.EventManager,
		groupMgr:       manager.GroupManager,
		applicationMgr: manager.ApplicationManager,
		clusterMgr:     manager.ClusterMgr,
		userMgr:        manager.UserManager,
	}
}

type messageDependency struct {
	webhook     *webhookmodels.Webhook
	event       *models.Event
	application *applicationmodels.Application
	cluster     *clustermodels.Cluster
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

// listAssociatedResources list all the associated resources of event to find all the webhooks
func (w *WebhookLogGenerator) listAssociatedResources(ctx context.Context,
	e *models.Event) (*messageDependency, map[string][]uint) {
	var (
		resources   map[string][]uint
		cluster     *clustermodels.Cluster
		application *applicationmodels.Application
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
	user, err := w.userMgr.GetUserByID(ctx, dep.event.CreatedBy)
	if err != nil {
		return "", err
	}

	message := MessageContent{
		EventID:   dep.event.ID,
		WebhookID: dep.webhook.ID,
		EventType: string(dep.event.EventType),
		User:      usermodels.ToUser(user),
		Extra:     dep.event.EventSummary.Extra,
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
			// 3.2 add webhook to the list
			if _, ok := conditionsToCreate[event.ID]; !ok {
				conditionsToCreate[event.ID] = map[uint]messageDependency{}
			}
			conditionsToCreate[event.ID][webhook.ID] = messageDependency{
				webhook:     webhook,
				event:       event,
				application: dependency.application,
				cluster:     dependency.cluster,
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
