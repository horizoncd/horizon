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

package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/webhook/models"
)

type DAO interface {
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
	ListWebhookLogsByStatus(ctx context.Context, wID uint,
		status string) ([]*models.WebhookLog, error)
	ListWebhookLogsByMap(ctx context.Context,
		webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error)
	UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error)
	GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error)
	GetWebhookLogByEventID(ctx context.Context, webhookID, eventID uint) (*models.WebhookLog, error)
	GetMaxEventIDOfLog(ctx context.Context) (uint, error)
	DeleteWebhookLogs(ctx context.Context, id ...uint) (int64, error)
	ListWebhookLogsForClean(ctx context.Context, query *q.Query) ([]*models.WebhookLogWithEventInfo, error)
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) CreateWebhook(ctx context.Context, webhook *models.Webhook) (*models.Webhook, error) {
	if result := d.db.WithContext(ctx).Create(webhook); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.WebhookInDB, result.Error.Error())
	}
	return webhook, nil
}

func (d *dao) GetWebhook(ctx context.Context, id uint) (*models.Webhook, error) {
	var w models.Webhook
	if result := d.db.WithContext(ctx).First(&w, id); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.WebhookInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.WebhookInDB, result.Error.Error())
	}
	return &w, nil
}

func (d *dao) ListWebhooks(ctx context.Context) ([]*models.Webhook, error) {
	var ws []*models.Webhook
	if result := d.db.WithContext(ctx).Find(&ws); result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.WebhookInDB, result.Error.Error())
	}
	return ws, nil
}

func (d *dao) ListWebhookOfResources(ctx context.Context,
	resources map[string][]uint, query *q.Query) ([]*models.Webhook, int64, error) {
	var ws []*models.Webhook
	var (
		condition *gorm.DB
		limit     int
		offset    int
		count     int64
	)
	// assemble condition
	for resourceType, resourceIDs := range resources {
		subCondition := d.db.Where("resource_type = ?", resourceType).
			Where("resource_id in ?", resourceIDs)
		if condition != nil {
			condition.Or(subCondition)
		} else {
			condition = subCondition
		}
	}

	statement := d.db.WithContext(ctx).
		Where(condition).
		Order("created_at desc").
		Limit(limit).
		Offset(offset)

	if query != nil {
		if v, ok := query.Keywords[common.CreatedAt]; ok {
			statement = statement.Where("created_at <= ?", v)
		}
		if v, ok := query.Keywords[common.Enabled]; ok {
			statement = statement.Where("enabled = ?", v)
		}
	}

	if result := statement.
		Find(&ws).
		Offset(-1).
		Count(&count); result.Error != nil {
		return nil, count, herrors.NewErrGetFailed(herrors.WebhookInDB, result.Error.Error())
	}
	return ws, count, nil
}

func (d *dao) UpdateWebhook(ctx context.Context, id uint,
	w *models.Webhook) (*models.Webhook, error) {
	if result := d.db.WithContext(ctx).Where("id = ?", id).
		Select("enabled", "url", "enable_ssl_verify", "description", "secret", "triggers").
		Updates(w); result.Error != nil {
		return nil, herrors.NewErrUpdateFailed(herrors.WebhookInDB, result.Error.Error())
	}
	return w, nil
}

func (d *dao) DeleteWebhook(ctx context.Context, id uint) error {
	deleteFunc := func(tx *gorm.DB) error {
		if result := d.db.WithContext(ctx).Where("webhook_id = ?", id).
			Delete(&models.WebhookLog{}); result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.WebhookInDB, result.Error.Error())
		}

		if result := d.db.WithContext(ctx).Delete(&models.Webhook{}, id); result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.WebhookInDB, result.Error.Error())
		}

		return nil
	}
	return d.db.WithContext(ctx).Transaction(deleteFunc)
}

func (d *dao) CreateWebhookLog(ctx context.Context,
	wl *models.WebhookLog) (*models.WebhookLog, error) {
	d.db.WithContext(ctx).Commit().Callback()
	if result := d.db.WithContext(ctx).Create(wl); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return wl, nil
}

func (d *dao) CreateWebhookLogs(ctx context.Context,
	wls []*models.WebhookLog) ([]*models.WebhookLog, error) {
	d.db.WithContext(ctx).Commit().Callback()
	if result := d.db.WithContext(ctx).Create(wls); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return wls, nil
}

func (d *dao) ListWebhookLogs(ctx context.Context, query *q.Query,
	resources map[string][]uint) ([]*models.WebhookLogWithEventInfo, int64, error) {
	var (
		logs  []*models.WebhookLogWithEventInfo
		count int64
	)

	stm := d.db.WithContext(ctx).Table("tb_webhook_log l").
		Joins("left join tb_event e on l.event_id=e.id").
		Select("l.*, e.resource_type, e.resource_id, e.event_type")

	if query != nil && query.Keywords != nil {
		for k, v := range query.Keywords {
			switch k {
			case common.Orphaned:
				stm = stm.Where("e.id is null")
			case common.WebhookID:
				stm = stm.Where("l.l.webhook_id = ?", v)
			case common.EventType:
				stm = stm.Where("e.event_type = ?", v)
			case common.Offset:
				if offset, ok := v.(int); ok {
					stm = stm.Offset(offset)
				}
			case common.Limit:
				if limit, ok := v.(int); ok {
					stm = stm.Limit(limit)
				}
			case common.StartID:
				stm = stm.Where("l.id > ?", v)
			case common.EndID:
				stm = stm.Where("l.id <= ?", v)
			case common.OrderBy:
				stm = stm.Order(v)
			}
		}
	}

	if len(resources) > 0 {
		var resourceCondition *gorm.DB
		for resourceType, resourceIDs := range resources {
			if resourceCondition == nil {
				resourceCondition = d.db.WithContext(context.Background()).
					Where("e.resource_type = ? and e.resource_id in ?", resourceType, resourceIDs)
			}
			resourceCondition = resourceCondition.
				Or("e.resource_type = ? and e.resource_id in ?", resourceType, resourceIDs)
		}
		stm.Where(resourceCondition)
	}

	if result := stm.
		Scan(&logs).
		Offset(-1).
		Count(&count); result.Error != nil {
		return nil, 0, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}

	return logs, count, nil
}

func (d *dao) ListWebhookLogsByMap(ctx context.Context,
	webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error) {
	var (
		ws        []*models.WebhookLog
		condition *gorm.DB
	)
	if len(webhookEventMap) == 0 {
		return nil, nil
	}
	// assemble condition
	for webhookID, eventIDs := range webhookEventMap {
		subCondition := d.db.Where("webhook_id = ?", webhookID).
			Where("event_id in ?", eventIDs)
		if condition != nil {
			condition.Or(subCondition)
		} else {
			condition = subCondition
		}
	}
	if result := d.db.WithContext(ctx).Where(condition).
		Find(&ws); result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return ws, nil
}

func (d *dao) ListWebhookLogsByStatus(ctx context.Context, wID uint,
	status string) ([]*models.WebhookLog, error) {
	var ws []*models.WebhookLog
	if result := d.db.WithContext(ctx).Where("webhook_id = ?", wID).Where("status = ?", status).
		Find(&ws); result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return ws, nil
}

func (d *dao) UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error) {
	if result := d.db.WithContext(ctx).Where("id = ?", wl.ID).
		Select("status", "response_headers", "response_body",
			"status", "error_message").
		Updates(wl); result.Error != nil {
		return nil, herrors.NewErrUpdateFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return wl, nil
}

func (d *dao) GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error) {
	var wl models.WebhookLog
	if result := d.db.WithContext(ctx).Where("id = ?", id).First(&wl); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.WebhookLogInDB, result.Error.Error()),
				fmt.Sprintf("failed to find webhook log by id: %d", id))
		}
		return nil, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return &wl, nil
}

func (d *dao) GetWebhookLogByEventID(ctx context.Context, webhookID, eventID uint) (*models.WebhookLog, error) {
	var wl models.WebhookLog
	if result := d.db.WithContext(ctx).
		Where("webhook_id = ?", webhookID).
		Where("event_id = ?", eventID).
		First(&wl); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.WebhookLogInDB, result.Error.Error()),
				fmt.Sprintf("failed to find webhook log by webhook id: %d, event id: %d",
					webhookID, eventID))
		}
		return nil, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return &wl, nil
}

func (d *dao) DeleteWebhookLogs(ctx context.Context, ids ...uint) (int64, error) {
	result := d.db.WithContext(ctx).Where("id in (?)", ids).Delete(&models.WebhookLog{})
	if result.Error != nil {
		return 0, herrors.NewErrDeleteFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return result.RowsAffected, nil
}

func (d *dao) GetMaxEventIDOfLog(ctx context.Context) (uint, error) {
	var maxID uint
	if result := d.db.WithContext(ctx).Model(&models.WebhookLog{}).Select("ifnull(max(event_id),0)").
		Scan(&maxID); result.Error != nil {
		return maxID, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return maxID, nil
}

func (d *dao) ListWebhookLogsForClean(ctx context.Context, query *q.Query) ([]*models.WebhookLogWithEventInfo, error) {
	var logs []*models.WebhookLogWithEventInfo

	statement := d.db.WithContext(ctx).Table("tb_webhook_log l").
		Joins("left join tb_event e on l.event_id=e.id").
		Select("l.id, l.updated_at, e.resource_type, e.resource_id, e.event_type")

	if query != nil {
		if v, ok := query.Keywords[common.StartID]; ok {
			statement = statement.Where("l.id > ?", v)
		}
		if v, ok := query.Keywords[common.Limit]; ok {
			if limit, ok := v.(int); ok {
				statement = statement.Limit(limit)
			}
		}
	}

	if result := statement.Find(&logs); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.WebhookLogInDB, result.Error.Error())
	}

	return logs, nil
}
