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

	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type WebhookManager interface {
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
}

type webhookManager struct {
	dao dao.WebhookDAO
}

func NewWebhookManager(db *gorm.DB) WebhookManager {
	return &webhookManager{
		dao: dao.NewWebhookDAO(db),
	}
}

func (m *webhookManager) CreateWebhook(ctx context.Context, w *models.Webhook) (*models.Webhook, error) {
	const op = "webhook webhookManager: create webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhook(ctx, w)
}

func (m *webhookManager) GetWebhook(ctx context.Context, id uint) (*models.Webhook, error) {
	const op = "webhook webhookManager: get webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhook(ctx, id)
}

func (m *webhookManager) ListWebhookOfResources(ctx context.Context,
	resources map[string][]uint, query *q.Query) ([]*models.Webhook, int64, error) {
	const op = "webhook webhookManager: list webhook of resources"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookOfResources(ctx, resources, query)
}

func (m *webhookManager) ListWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	const op = "webhook webhookManager: list webhooks"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhooks(ctx)
}

func (m *webhookManager) UpdateWebhook(ctx context.Context, id uint,
	w *models.Webhook) (*models.Webhook, error) {
	const op = "webhook webhookManager: update webhook"
	defer wlog.Start(ctx, op).StopPrint()

	return m.dao.UpdateWebhook(ctx, id, w)
}

func (m *webhookManager) DeleteWebhook(ctx context.Context, id uint) error {
	const op = "webhook webhookManager: delete webhook"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.DeleteWebhook(ctx, id)
}

func (m *webhookManager) CreateWebhookLog(ctx context.Context,
	wl *models.WebhookLog) (*models.WebhookLog, error) {
	const op = "webhook webhookManager: create webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhookLog(ctx, wl)
}

func (m *webhookManager) ListWebhookLogs(ctx context.Context, query *q.Query,
	resources map[string][]uint) ([]*models.WebhookLogWithEventInfo, int64, error) {
	const op = "webhook manager: list webhook logs"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookLogs(ctx, query, resources)
}

func (m *webhookManager) ListWebhookLogsByMap(ctx context.Context,
	webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error) {
	const op = "webhook webhookManager: list webhook logs by webhooks and events map"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListWebhookLogsByMap(ctx, webhookEventMap)
}

func (m *webhookManager) CreateWebhookLogs(ctx context.Context,
	wls []*models.WebhookLog) ([]*models.WebhookLog, error) {
	const op = "webhook webhookManager: create webhook logs"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateWebhookLogs(ctx, wls)
}

func (m *webhookManager) ListWebhookLogsByStatus(ctx context.Context, wID uint,
	status string) ([]*models.WebhookLog, error) {
	return m.dao.ListWebhookLogsByStatus(ctx, wID, status)
}

func (m *webhookManager) UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error) {
	const op = "webhook webhookManager: update  webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.UpdateWebhookLog(ctx, wl)
}

func (m *webhookManager) GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error) {
	const op = "webhook webhookManager: get webhook log"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhookLog(ctx, id)
}

func (m *webhookManager) GetWebhookLogByEventID(ctx context.Context,
	webhookID, eventID uint) (*models.WebhookLog, error) {
	const op = "webhook webhookManager: get webhook log by event id"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetWebhookLogByEventID(ctx, webhookID, eventID)
}

func (m *webhookManager) DeleteWebhookLogs(ctx context.Context, id ...uint) (int64, error) {
	const op = "webhook manager: delete webhook log"
	defer wlog.Start(ctx, op).StopPrint()

	return m.dao.DeleteWebhookLogs(ctx, id...)
}

func (m *webhookManager) ResendWebhook(ctx context.Context, id uint) (*models.WebhookLog, error) {
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
		Status:          models.WebhookStatusWaiting,
	}
	return m.dao.CreateWebhookLog(ctx, &wlCopy)
}

func (m *webhookManager) GetMaxEventIDOfLog(ctx context.Context) (uint, error) {
	return m.dao.GetMaxEventIDOfLog(ctx)
}
