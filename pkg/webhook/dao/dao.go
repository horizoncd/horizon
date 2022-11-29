package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/webhook/models"
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
	ListWebhookLogs(ctx context.Context, wID uint, query *q.Query) ([]*models.WebhookLogWithEventInfo, int64, error)
	ListWebhookLogsByStatus(ctx context.Context, wID uint,
		status string) ([]*models.WebhookLog, error)
	ListWebhookLogsByMap(ctx context.Context,
		webhookEventMap map[uint][]uint) ([]*models.WebhookLog, error)
	UpdateWebhookLog(ctx context.Context, wl *models.WebhookLog) (*models.WebhookLog, error)
	GetWebhookLog(ctx context.Context, id uint) (*models.WebhookLog, error)
	GetWebhookLogByEventID(ctx context.Context, webhookID, eventID uint) (*models.WebhookLog, error)
	GetMaxEventIDOfLog(ctx context.Context) (uint, error)
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

	if result := d.db.WithContext(ctx).
		Where(condition).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&ws).
		Offset(-1).
		Limit(-1).
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

func (d *dao) ListWebhookLogs(ctx context.Context, wID uint,
	query *q.Query) ([]*models.WebhookLogWithEventInfo, int64, error) {
	var (
		logs  []*models.WebhookLogWithEventInfo
		count int64
	)

	getStatementByResource := func(resourceType, tableName string) *gorm.DB {
		return d.db.Table("tb_webhook_log l").
			Joins("join tb_event e on l.event_id=e.id").
			Joins(fmt.Sprintf("join %s r on e.resource_id=r.id", tableName)).
			Where("l.webhook_id=?", wID).
			Where("e.resource_type=?", resourceType).
			Select("l.*, e.resource_type, e.resource_id, e.action, r.name as resource_name")
	}

	statement := d.db.Raw("? union ? union ?",
		getStatementByResource(common.ResourceGroup, "tb_group"),
		getStatementByResource(common.ResourceApplication, "tb_application"),
		getStatementByResource(common.ResourceCluster, "tb_cluster"),
	)

	if result := d.db.WithContext(ctx).Raw("select *"+
		" from (?) as w order by created_at desc limit ? offset ?",
		statement, query.Limit(), query.Offset()).Scan(&logs); result.Error != nil {
		return nil, count, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}

	if result := d.db.WithContext(ctx).Raw("select count(1) from (?) as w",
		statement).Scan(&count); result.Error != nil {
		return nil, count, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
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

func (d *dao) GetMaxEventIDOfLog(ctx context.Context) (uint, error) {
	var maxID uint
	if result := d.db.WithContext(ctx).Model(&models.WebhookLog{}).Select("ifnull(max(event_id),0)").
		Scan(&maxID); result.Error != nil {
		return maxID, herrors.NewErrGetFailed(herrors.WebhookLogInDB, result.Error.Error())
	}
	return maxID, nil
}
