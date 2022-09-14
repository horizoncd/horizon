package dao

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/event/models"
)

type DAO interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error)
	ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error)
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

func (d *dao) ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error) {
	var events []*models.Event
	if result := d.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&events); result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EventInDB, result.Error.Error())
	}
	return events, nil
}

func (d *dao) ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error) {
	var events []*models.Event
	if result := d.db.WithContext(ctx).Order("id asc").Where("id >= ?", start).
		Where("id <= ?", end).Find(&events); result.Error != nil {
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
		return nil, herrors.NewErrInsertFailed(herrors.EventInDB, result.Error.Error())
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
		return nil, herrors.NewErrInsertFailed(herrors.EventCursorInDB, result.Error.Error())
	}
	return &eventIndex, nil
}

// todo: must add gc
