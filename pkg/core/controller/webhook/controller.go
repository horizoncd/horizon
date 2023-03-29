package webhook

import (
	"context"

	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/lib/q"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/param"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	wmanager "github.com/horizoncd/horizon/pkg/webhook/manager"
	"github.com/horizoncd/horizon/pkg/webhook/models"
)

type Controller interface {
	CreateWebhook(ctx context.Context, resourceType string,
		resourceID uint, w *CreateWebhookRequest) (*Webhook, error)
	GetWebhook(ctx context.Context, id uint) (*Webhook, error)
	ListWebhooks(ctx context.Context, resourceType string,
		resourceID uint, query *q.Query) ([]*Webhook, int64, error)
	UpdateWebhook(ctx context.Context, id uint,
		w *UpdateWebhookRequest) (*Webhook, error)
	DeleteWebhook(ctx context.Context, id uint) error
	ListWebhookLogs(ctx context.Context, wID uint, query *q.Query) ([]*LogSummary, int64, error)
	GetWebhookLog(ctx context.Context, id uint) (*Log, error)
	ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error)
}

type controller struct {
	webhookMgr     wmanager.Manager
	userMgr        usermanager.Manager
	eventMgr       eventmanager.Manager
	groupMgr       groupmanager.Manager
	applicationMgr applicationmanager.Manager
	clusterMgr     clustermanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		webhookMgr:     param.WebhookManager,
		userMgr:        param.UserManager,
		eventMgr:       param.EventManager,
		clusterMgr:     param.ClusterMgr,
		applicationMgr: param.ApplicationManager,
		groupMgr:       param.GroupManager,
	}
}

func (c *controller) CreateWebhook(ctx context.Context, resourceType string,
	resourceID uint, w *CreateWebhookRequest) (*Webhook, error) {
	const op = "webhook controller: create"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. validate request
	if err := c.validateCreateRequest(resourceType, w); err != nil {
		return nil, err
	}

	// 2. transfer model
	wm, err := w.toModel(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	// 3. create webhook
	wm, err = c.webhookMgr.CreateWebhook(ctx, wm)
	if err != nil {
		return nil, err
	}

	return ofWebhookModel(wm), nil
}

func (c *controller) GetWebhook(ctx context.Context, id uint) (*Webhook, error) {
	const op = "wehook controller: get"
	defer wlog.Start(ctx, op).StopPrint()

	w, err := c.webhookMgr.GetWebhook(ctx, id)
	if err != nil {
		return nil, err
	}
	return ofWebhookModel(w), nil
}

func (c *controller) UpdateWebhook(ctx context.Context, id uint,
	w *UpdateWebhookRequest) (*Webhook, error) {
	const op = "wehook controller: update"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. validate request
	if err := c.validateUpdateRequest(w); err != nil {
		return nil, err
	}

	// 2. transfer model
	wm, err := c.webhookMgr.GetWebhook(ctx, id)
	if err != nil {
		return nil, err
	}
	wm = w.toModel(wm)

	// 3. update webhook
	wm, err = c.webhookMgr.UpdateWebhook(ctx, id, wm)
	if err != nil {
		return nil, err
	}

	return ofWebhookModel(wm), nil
}

func (c *controller) DeleteWebhook(ctx context.Context, id uint) error {
	const op = "wehook controller: delete"
	defer wlog.Start(ctx, op).StopPrint()

	return c.webhookMgr.DeleteWebhook(ctx, id)
}

func (c *controller) ListWebhooks(ctx context.Context, resourceType string,
	resourceID uint, query *q.Query) ([]*Webhook, int64, error) {
	const op = "wehook controller: list"
	defer wlog.Start(ctx, op).StopPrint()

	resource := map[string][]uint{
		resourceType: {resourceID},
	}
	webhooks, total, err := c.webhookMgr.ListWebhookOfResources(ctx, resource, query)
	if err != nil {
		return nil, total, err
	}

	var ws []*Webhook
	for _, w := range webhooks {
		ws = append(ws, ofWebhookModel(w))
	}
	return ws, total, nil
}

func (c *controller) ListWebhookLogs(ctx context.Context, wID uint,
	query *q.Query) ([]*LogSummary, int64, error) {
	const op = "wehook controller: list log"
	defer wlog.Start(ctx, op).StopPrint()

	resources := map[string][]uint{}

	if filter, ok := query.Keywords[common.Filter].(string); ok {
		groups, err := c.groupMgr.GetByNameFuzzilyIncludeSoftDelete(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
		for _, group := range groups {
			resources[common.ResourceGroup] = append(resources[common.ResourceGroup], group.ID)
		}

		apps, err := c.applicationMgr.GetByNameFuzzilyIncludeSoftDelete(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
		for _, app := range apps {
			resources[common.ResourceApplication] = append(resources[common.ResourceGroup], app.ID)
		}

		clusters, err := c.clusterMgr.GetByNameFuzzilyIncludeSoftDelete(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
		for _, cluster := range clusters {
			resources[common.ResourceCluster] = append(resources[common.ResourceCluster], cluster.ID)
		}

		if len(resources) == 0 {
			return nil, 0, nil
		}
	}

	webhookLogs, total, err := c.webhookMgr.ListWebhookLogs(ctx, wID, query, resources)
	if err != nil {
		return nil, total, err
	}

	var wls []*LogSummary
	for _, wl := range webhookLogs {
		switch wl.ResourceType {
		case common.ResourceApplication:
			application, err := c.applicationMgr.GetByIDIncludeSoftDelete(ctx, wl.ResourceID)
			if err != nil {
				return nil, 0, err
			}
			wl.ResourceName = application.Name
		case common.ResourceCluster:
			cluster, err := c.clusterMgr.GetByIDIncludeSoftDelete(ctx, wl.ResourceID)
			if err != nil {
				return nil, 0, err
			}
			wl.ResourceName = cluster.Name
		}
		wls = append(wls, ofWebhookLogSummaryModel(wl))
	}
	return wls, total, nil
}

func (c *controller) GetWebhookLog(ctx context.Context, id uint) (*Log, error) {
	const op = "wehook controller: get log"
	defer wlog.Start(ctx, op).StopPrint()

	wl, err := c.webhookMgr.GetWebhookLog(ctx, id)
	if err != nil {
		return nil, err
	}

	userMap, err := c.userMgr.GetUserMapByIDs(ctx,
		[]uint{wl.CreatedBy})
	if err != nil {
		return nil, err
	}

	webhookLog := ofWebhookLogModel(wl)
	webhookLog.CreatedBy = usermodels.ToUser(userMap[wl.CreatedBy])
	return webhookLog, nil
}

func (c *controller) ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error) {
	const op = "wehook controller: resend"
	defer wlog.Start(ctx, op).StopPrint()

	return c.webhookMgr.ResendWebhook(ctx, id)
}
