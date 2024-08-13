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

package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/horizoncd/horizon/pkg/webhook/dao"
	models "github.com/horizoncd/horizon/pkg/webhook/models"
)

type Manager interface {
	CreateWebhook(ctx context.Context, webhook *models.Webhook) (*models.Webhook, error)
	GetWebhook(ctx context.Context, id uint) (*models.Webhook, error)
	ListWebhookOfResources(ctx context.Context,
		resources map[string][]uint, query *q.Query) ([]*models.Webhook, int64, error)
	ListWebhooks(ctx context.Context) ([]*models.Webhook, error)
	UpdateWebhook(ctx context.Context, id uint, w *models.Webhook) (*models.Webhook, error)
	DeleteWebhook(ctx context.Context, id uint) error
	CreateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error)
	CreateWebhookLogs(ctx context.Context, wls []*models.WebhookLog) ([]*models.WebhookLog, error)
	ListWebhookLogs(ctx context.Context, query *q.Query,
		resources map[string][]uint) ([]*models.WebhookLogWithEventInfo, int64, error)
	ListWebhookLogsByMap(ctx context.Context,
		webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error)
	ListWebhookLogsByStatus(ctx context.Context, wID uint,
		status string) ([]*models.WebhookLog, error)
	UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error)
	GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error)
	ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error)
	GetWebhookLogByEventID(ctx context.Context, webhookID, eventID uint) (*models.WebhookLog, error)
	GetMaxEventIDOfLog(ctx context.Context) (uint, error)
	DeleteWebhookLogs(ctx context.Context, id ...uint) (int64, error)
	ListWebhookLogsForClean(ctx context.Context, query *q.Query) ([]*models.WebhookLogWithEventInfo, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) CreateWebhook(ctx context.Context, w *models.Webhook) (*models.Webhook, error) {
	const op = "webhook manager: create webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhook(ctx, w)
}

func (m *manager) GetWebhook(ctx context.Context, id uint) (*models.Webhook, error) {
	const op = "webhook manager: get webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhook(ctx, id)
}

func (m *manager) ListWebhookOfResources(ctx context.Context,
	resources map[string][]uint, query *q.Query) ([]*models.Webhook, int64, error) {
	const op = "webhook manager: list webhook of resources"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookOfResources(ctx, resources, query)
}

func (m *manager) ListWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	const op = "webhook manager: list webhooks"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhooks(ctx)
}

func (m *manager) UpdateWebhook(ctx context.Context, id uint,
	w *models.Webhook) (*models.Webhook, error) {
	const op = "webhook manager: update webhook"
	defer wlog.Start(ctx, op).StopPrint()

	return m.dao.UpdateWebhook(ctx, id, w)
}

func (m *manager) DeleteWebhook(ctx context.Context, id uint) error {
	const op = "webhook manager: delete webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.DeleteWebhook(ctx, id)
}

func (m *manager) CreateWebhookLog(ctx context.Context,
	wl *models.WebhookLog) (*models.WebhookLog, error) {
	const op = "webhook manager: create webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhookLog(ctx, wl)
}

func (m *manager) ListWebhookLogs(ctx context.Context, query *q.Query,
	resources map[string][]uint) ([]*models.WebhookLogWithEventInfo, int64, error) {
	const op = "webhook manager: list webhook logs"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookLogs(ctx, query, resources)
}

func (m *manager) ListWebhookLogsByMap(ctx context.Context,
	webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error) {
	const op = "webhook manager: list webhook logs by webhooks and events map"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookLogsByMap(ctx, webhookEventMap)
}

func (m *manager) CreateWebhookLogs(ctx context.Context, wls []*models.WebhookLog) ([]*models.WebhookLog, error) {
	const op = "webhook manager: create webhook logs"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhookLogs(ctx, wls)
}

func (m *manager) ListWebhookLogsByStatus(ctx context.Context, wID uint,
	status string) ([]*models.WebhookLog, error) {
	return m.dao.ListWebhookLogsByStatus(ctx, wID, status)
}

func (m *manager) UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error) {
	const op = "webhook manager: update  webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.UpdateWebhookLog(ctx, wl)
}

func (m *manager) GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error) {
	const op = "webhook manager: get webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhookLog(ctx, id)
}

func (m *manager) GetWebhookLogByEventID(ctx context.Context, webhookID, eventID uint) (*models.WebhookLog, error) {
	const op = "webhook manager: get webhook log by event id"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhookLogByEventID(ctx, webhookID, eventID)
}

func (m *manager) DeleteWebhookLogs(ctx context.Context, id ...uint) (int64, error) {
	const op = "webhook manager: delete webhook log"
	defer wlog.Start(ctx, op).StopPrint()

	return m.dao.DeleteWebhookLogs(ctx, id...)
}

func (m *manager) ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error) {
	const op = "webhook manager: resend"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get webhook log
	wl, err := m.dao.GetWebhookLog(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. make a copy with waiting status
	wlCopy := models.WebhookLog{
		WebhookID:       wl.WebhookID,
		EventID:         wl.EventID,
		URL:             wl.URL,
		RequestHeaders:  wl.RequestHeaders,
		RequestData:     wl.RequestData,
		ResponseHeaders: wl.ResponseHeaders,
		ResponseBody:    wl.ResponseBody,
		Status:          models.StatusWaiting,
	}
	return m.dao.CreateWebhookLog(ctx, &wlCopy)
}

func (m *manager) GetMaxEventIDOfLog(ctx context.Context) (uint, error) {
	return m.dao.GetMaxEventIDOfLog(ctx)
}

func (m *manager) ListWebhookLogsForClean(
	ctx context.Context,
	query *q.Query,
) ([]*models.WebhookLogWithEventInfo, error) {
	const op = "webhook manager: list webhook logs for clean"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookLogsForClean(ctx, query)
}
