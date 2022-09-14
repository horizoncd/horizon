package wlgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-yaml/yaml"

	"g.hz.netease.com/horizon/core/common"
	webhookctl "g.hz.netease.com/horizon/core/controller/webhook"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	"g.hz.netease.com/horizon/pkg/event/models"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	webhookmanager "g.hz.netease.com/horizon/pkg/webhook/manager"
	webhookmodels "g.hz.netease.com/horizon/pkg/webhook/models"
)

const (
	WebhookSecretHeader      = "X-Horizon-Webhook-Secret"
	WebhookContentTypeHeader = "Content-Type"
	WebhookContentType       = "application/json;charset=utf-8"
)

type MessageContent struct {
	ID           uint             `json:"id,omitempty"`
	EventID      uint             `json:"eventID,omitempty"`
	WebhookID    uint             `json:"webhookID,omitempty"`
	ResourceType string           `json:"resourceType,omitempty"`
	Application  *ApplicationInfo `json:"application,omitempty"`
	Cluster      *ClusterInfo     `json:"cluster,omitempty"`
	Action       string           `json:"action,omitempty"`
	User         *common.User     `json:"user,omitempty"`
}

type ResourceCommonInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ApplicationInfo struct {
	ResourceCommonInfo
	Priority string `json:"priority,omitempty"`
}

type ClusterInfo struct {
	ResourceCommonInfo
	ApplicationName string `json:"applicationName,omitempty"`
	Env             string `json:"env,omitempty"`
}

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

func (w *WebhookLogGenerator) Process(ctx context.Context, events []*models.Event,
	resume bool) error {
	listSystemResources := func() map[string][]uint {
		return map[string][]uint{
			string(models.Group): {0},
		}
	}

	type messageDependency struct {
		webhook     *webhookmodels.Webhook
		event       *models.Event
		application *applicationmodels.Application
		cluster     *clustermodels.Cluster
	}

	listAssociatedResourcesOfApp := func(id uint) (*applicationmodels.Application, map[string][]uint) {
		resources := listSystemResources()
		app, err := w.applicationMgr.GetByID(ctx, id)
		if err != nil {
			log.Warningf(ctx, "application %d is not exist", id)
			return nil, resources
		}
		resources[string(models.Application)] = []uint{app.ID}

		group, err := w.groupMgr.GetByID(ctx, app.GroupID)
		if err != nil {
			log.Warningf(ctx, "application %d is not exist", id)
			return app, resources
		}
		groupIDs := groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs)
		resources[string(models.Group)] = append(resources[string(models.Group)], groupIDs...)
		return app, resources
	}

	listAssociatedResourcesOfCluster := func(id uint) (*clustermodels.Cluster,
		*applicationmodels.Application, map[string][]uint) {
		cluster, err := w.clusterMgr.GetByID(ctx, id)
		if err != nil {
			log.Warningf(ctx, "cluster %d is not exist",
				cluster.ID)
			return nil, nil, nil
		}
		app, resources := listAssociatedResourcesOfApp(cluster.ApplicationID)
		if resources == nil {
			resources = map[string][]uint{}
		}
		resources[string(models.Cluster)] = []uint{cluster.ID}
		return cluster, app, resources
	}

	listAssociatedResources := func(e *models.Event) (*messageDependency, map[string][]uint) {
		var (
			resources   map[string][]uint
			cluster     *clustermodels.Cluster
			application *applicationmodels.Application
			dep         = &messageDependency{}
		)

		switch e.ResourceType {
		case models.Application:
			application, resources = listAssociatedResourcesOfApp(e.ResourceID)
			dep.application = application
		case models.Cluster:
			cluster, application, resources = listAssociatedResourcesOfCluster(e.ResourceID)
			dep.application = application
			dep.cluster = cluster
		default:
			log.Infof(ctx, "resource type %s is unsupported",
				e.ResourceType)
		}
		return dep, resources
	}

	makeHeaders := func(secret string) (string, error) {
		header := http.Header{}
		header.Add(WebhookSecretHeader, secret)
		header.Add(WebhookContentTypeHeader, WebhookContentType)
		headerByte, err := yaml.Marshal(header)
		if err != nil {
			return "", err
		}
		return string(headerByte), nil
	}

	makeBody := func(dep *messageDependency) (string, error) {
		user, err := w.userMgr.GetUserByID(ctx, dep.event.CreatedBy)
		if err != nil {
			return "", err
		}

		message := MessageContent{
			EventID:      dep.event.ID,
			WebhookID:    dep.webhook.ID,
			ResourceType: string(dep.event.ResourceType),
			Action:       string(dep.event.Action),
			User:         common.ToUser(user),
		}

		if dep.application != nil {
			message.Application = &ApplicationInfo{
				ResourceCommonInfo: ResourceCommonInfo{
					ID:   dep.application.ID,
					Name: dep.application.Name,
				},
				Priority: string(dep.application.Priority),
			}
		}

		if dep.cluster != nil {
			message.Cluster = &ClusterInfo{
				ResourceCommonInfo: ResourceCommonInfo{
					ID:   dep.application.ID,
					Name: dep.application.Name,
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

	var (
		webhookLogs        []*webhookmodels.WebhookLog
		conditionsToCreate = map[uint]map[uint]messageDependency{}
		conditionsToQuery  = map[uint][]uint{}
	)

	for _, event := range events {
		// 1. get associated resources according to event resource type
		dependency, resources := listAssociatedResources(event)
		if resources == nil {
			continue
		}

		// 2. list webhooks of all associated resources
		webhooks, _, err := w.webhookMgr.ListWebhookOfResources(ctx, resources, nil)
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
			headers, err := makeHeaders(dependency.webhook.Secret)
			if err != nil {
				log.Errorf(ctx, fmt.Sprintf("failed to make headers, error: %+v", err))
				continue
			}

			body, err := makeBody(&dependency)
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
	}
	return nil
}
