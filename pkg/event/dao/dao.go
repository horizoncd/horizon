package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/event/models"
)

type DAO interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	ListEvents(ctx context.Context, query *q.Query) ([]*models.Event, error)
	CreateOrUpdateCursor(ctx context.Context,
		eventIndex *models.EventCursor) (*models.EventCursor, error)
	GetCursor(ctx context.Context) (*models.EventCursor, error)
	GetEvent(ctx context.Context, id uint) (*models.Event, error)
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	if result := d.db.WithContext(ctx).Create(event); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EventInDB, result.Error.Error())
	}
	return event, nil
}

func (d *dao) ListEvents(ctx context.Context, query *q.Query) ([]*models.Event, error) {
	var events []*models.Event
	statement := d.db.WithContext(ctx).Order("id asc")
	for k, v := range query.Keywords {
		switch k {
		case common.Offset:
			offset, ok := v.(int)
			if !ok {
				return nil, perror.Wrap(herrors.ErrParamInvalid, fmt.Sprintf("invalid offset %v", v))
			}
			statement = statement.Offset(offset)
		case common.Limit:
			limit, ok := v.(int)
			if !ok {
				return nil, perror.Wrap(herrors.ErrParamInvalid, fmt.Sprintf("invalid limit %v", v))
			}
			statement = statement.Limit(limit)
		case common.StartID:
			statement = statement.Where("id >= ?", v)
		case common.EndID:
			statement = statement.Where("id <= ?", v)
		}
	}

	if result := statement.Find(&events); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EventInDB, result.Error.Error())
	}
	return events, nil
}

func (d *dao) GetEvent(ctx context.Context, id uint) (*models.Event, error) {
	var event *models.Event
	if result := d.db.WithContext(ctx).Where("id = ?", id).Find(&event); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EventInDB,
				result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EventInDB, result.Error.Error())
	}
	return event, nil
}

func (d *dao) CreateOrUpdateCursor(ctx context.Context,
	eventCursor *models.EventCursor) (*models.EventCursor, error) {
	if result := d.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "id",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"position"}),
	}).Create(eventCursor); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EventInDB, result.Error.Error())
	}
	return eventCursor, nil
}

func (d *dao) GetCursor(ctx context.Context) (*models.EventCursor, error) {
	var eventIndex models.EventCursor
	if result := d.db.First(&eventIndex); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EventCursorInDB,
				result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EventCursorInDB, result.Error.Error())
	}
	return &eventIndex, nil
}

// TODO: must add gc
