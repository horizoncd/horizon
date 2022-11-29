package manager

import (
	"context"

	"gorm.io/gorm"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/event/dao"
	"g.hz.netease.com/horizon/pkg/event/models"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Manager interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error)
	ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error)
	CreateOrUpdateCursor(ctx context.Context,
		eventIndex *models.EventCursor) (*models.EventCursor, error)
	GetCursor(ctx context.Context) (*models.EventCursor, error)
	GetEvent(ctx context.Context, id uint) (*models.Event, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) CreateEvent(ctx context.Context,
	event *models.Event) (*models.Event, error) {
	const op = "event manager: create event"
	defer wlog.Start(ctx, op).StopPrint()

	if event.ReqID == "" {
		rid, err := requestid.FromContext(ctx)
		if err != nil {
			log.Warningf(ctx, "failed to get request id, err: %s", err.Error())
		}
		event.ReqID = rid
	}
	e, err := m.dao.CreateEvent(ctx, event)
	if err != nil {
		return nil, herrors.NewErrCreateFailed(herrors.EventInDB, err.Error())
	}

	return e, nil
}

func (m *manager) CreateOrUpdateCursor(ctx context.Context,
	eventCursor *models.EventCursor) (*models.EventCursor, error) {
	const op = "event manager: create or update cursor"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.CreateOrUpdateCursor(ctx, eventCursor)
}

func (m *manager) GetCursor(ctx context.Context) (*models.EventCursor, error) {
	const op = "event manager: get cursor"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetCursor(ctx)
}

func (m *manager) ListEvents(ctx context.Context, offset, limit int) ([]*models.Event, error) {
	const op = "event manager: list events"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListEvents(ctx, offset, limit)
}

func (m *manager) ListEventsByRange(ctx context.Context, start, end uint) ([]*models.Event, error) {
	const op = "event manager: list events by range"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.ListEventsByRange(ctx, start, end)
}

func (m *manager) GetEvent(ctx context.Context, id uint) (*models.Event, error) {
	const op = "event manager: get event"
	defer wlog.Start(ctx, op).StopPrint()
	return m.dao.GetEvent(ctx, id)
}
