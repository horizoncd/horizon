package webhook

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	"g.hz.netease.com/horizon/pkg/param"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	wmanager "g.hz.netease.com/horizon/pkg/webhook/manager"
	"g.hz.netease.com/horizon/pkg/webhook/models"
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
	webhookMgr wmanager.Manager
	userMgr    usermanager.Manager
	eventMgr   eventmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		webhookMgr: param.WebhookManager,
		userMgr:    param.UserManager,
		eventMgr:   param.EventManager,
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
	wm = w.toModel(ctx, wm)

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

	webhookLogs, total, err := c.webhookMgr.ListWebhookLogs(ctx, wID, query)
	if err != nil {
		return nil, total, err
	}

	var wls []*LogSummary
	for _, wl := range webhookLogs {
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
	webhookLog.CreatedBy = common.ToUser(userMap[wl.CreatedBy])
	return webhookLog, nil
}

func (c *controller) ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error) {
	const op = "wehook controller: resend"
	defer wlog.Start(ctx, op).StopPrint()

	return c.webhookMgr.ResendWebhook(ctx, id)
}
