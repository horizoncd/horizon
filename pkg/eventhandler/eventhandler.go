package eventhandler

import (
	"context"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	"g.hz.netease.com/horizon/pkg/event/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/util/log"
	webhookmanager "g.hz.netease.com/horizon/pkg/webhook/manager"
)

type Service interface {
	RegisterEventHandler(name string, eh EventHandler) error
	StopAndWait()
	Start()
}

type cursor struct {
	ID       uint
	Position uint
}

type eventHandlerService struct {
	ctx           context.Context
	eventHandlers map[string]EventHandler
	cursor        *cursor
	resume        bool
	quit          chan bool

	eventMgr   eventmanager.Manager
	webhookMgr webhookmanager.Manager
}

func NewService(ctx context.Context, manager *managerparam.Manager) Service {
	return &eventHandlerService{
		ctx:           ctx,
		eventHandlers: map[string]EventHandler{},
		resume:        true,

		eventMgr:   manager.EventManager,
		webhookMgr: manager.WebhookManager,
	}
}

// EventHandler can be registered to process events
type EventHandler interface {
	Process(ctx context.Context, event []*models.Event, resume bool) error
}

func (e *eventHandlerService) RegisterEventHandler(name string, eh EventHandler) error {
	if _, ok := e.eventHandlers[name]; ok {
		return perror.Wrapf(herrors.ErrEventHandlerAlreadyExist, "%s already exist", name)
	}
	e.eventHandlers[name] = eh
	return nil
}

func (e *eventHandlerService) StopAndWait() {
	// 1. notify handlers to stop
	e.quit <- true
	// 2. wait for stop
	<-e.quit
	log.Info(e.ctx, "stop event handler queue")
}

func (e *eventHandlerService) Start() {
	go func() {
		// 1. get cursor
		for e.cursor == nil {
			eventCursor, err := e.eventMgr.GetCursor(e.ctx)
			if err != nil {
				if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
					log.Infof(e.ctx, "index does not exist, start process directly")
					e.cursor = &cursor{}
					break
				} else {
					log.Errorf(e.ctx, "failed to get event cursor, error: %+v", err)
					time.Sleep(time.Second * 3)
					continue
				}
			}
			e.cursor = &cursor{
				ID:       eventCursor.ID,
				Position: eventCursor.Position,
			}
		}

		// 2. process event
		limit := uint(5)
		saveCursorTicker := time.NewTicker(time.Second * 10)
		waitInterval := time.Second * 3
	L:
		for {
			select {
			case <-e.quit:
				log.Infof(e.ctx, "save cursor(%d) and stop event handlers", e.cursor.Position)
				e.saveCursor()
				close(e.quit)
				break L
			case <-saveCursorTicker.C:
				e.saveCursor()
			default:
				var (
					events []*models.Event
					err    error
				)

				if !e.resume {
					events, err = e.eventMgr.ListEvents(e.ctx, int(e.cursor.Position), int(limit))
					if err != nil {
						log.Errorf(e.ctx, "failed to list event by offset: %d, limit: %d",
							e.cursor.Position, limit)
						time.Sleep(waitInterval)
						continue
					}
					if len(events) == 0 {
						time.Sleep(waitInterval)
						continue
					}
				} else {
					// resume: continue to process the events that are halfway before restart
					lastProcessingCursor, err := e.getLastProcessingCursor()
					if err != nil {
						log.Error(e.ctx, "failed to get last processing cursor")
						time.Sleep(waitInterval)
						continue
					}
					if lastProcessingCursor <= e.cursor.Position {
						e.resume = false
						continue
					}
					events, err = e.eventMgr.ListEventsByRange(e.ctx, e.cursor.Position, lastProcessingCursor)
					if err != nil {
						log.Errorf(e.ctx, "failed to list event by limit %d offset %d", e.cursor.Position, limit)
						time.Sleep(waitInterval)
						continue
					}
				}
				for name, eh := range e.eventHandlers {
					if err := eh.Process(e.ctx, events, e.resume); err != nil {
						log.Errorf(e.ctx, "Failed to process event by handler %s, error: %s",
							name, err.Error())
						continue
					}
					e.cursor.Position = events[len(events)-1].ID
				}
				if e.resume {
					e.resume = false
				}
			}
		}
	}()
}

// getLastProcessingCursor get the last processing event cursor to resume
func (e *eventHandlerService) getLastProcessingCursor() (uint, error) {
	return e.webhookMgr.GetMaxEventIDOfLog(e.ctx)
}

func (e *eventHandlerService) saveCursor() {
	if _, err := e.eventMgr.CreateOrUpdateCursor(e.ctx, &models.EventCursor{
		ID:       e.cursor.ID,
		Position: e.cursor.Position,
	}); err != nil {
		log.Errorf(e.ctx, "failed to save cursor(%d), error: %+v",
			e.cursor.Position, err)
	}
}
