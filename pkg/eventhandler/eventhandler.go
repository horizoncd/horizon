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

package eventhandler

import (
	"context"
	"time"

	corecommon "github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	eventhandlerconfig "github.com/horizoncd/horizon/pkg/config/eventhandler"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/util/common"
	"github.com/horizoncd/horizon/pkg/util/log"
	webhookmanager "github.com/horizoncd/horizon/pkg/webhook/manager"
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
	config        eventhandlerconfig.Config
	ctx           context.Context
	eventHandlers map[string]EventHandler
	cursor        *cursor
	resume        bool
	quit          chan bool

	eventMgr   eventmanager.Manager
	webhookMgr webhookmanager.Manager
}

func NewService(ctx context.Context, manager *managerparam.Manager, config eventhandlerconfig.Config) Service {
	return &eventHandlerService{
		config:        config,
		ctx:           ctx,
		eventHandlers: map[string]EventHandler{},
		resume:        true,

		eventMgr:   manager.EventManager,
		webhookMgr: manager.WebhookManager,
	}
}

// EventHandler processes new events by registered handlers
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

// StopAndWait stop and wait for all registered handlers to exit
func (e *eventHandlerService) StopAndWait() {
	// 1. notify handlers to stop
	e.quit <- true
	// 2. wait for stop
	<-e.quit
	log.Info(e.ctx, "stop event handler queue")
}

func (e *eventHandlerService) Start() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf(e.ctx, "event handler service panic: %v", err)
				common.PrintStack()
			}
		}()

		// 1. get cursor
		for e.cursor == nil {
			eventCursor, err := e.getCursor(e.ctx)
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
		batchEventsCount := e.config.BatchEventsCount
		cursorSaveInterval := time.NewTicker(time.Second * time.Duration(e.config.CursorSaveInterval))
		idleWaitInterval := time.Second * time.Duration(e.config.IdleWaitInterval)
	L:
		for {
			select {
			case <-e.quit:
				log.Infof(e.ctx, "save cursor(%d) and stop event handlers", e.cursor.Position)
				e.saveCursor()
				close(e.quit)
				break L
			case <-cursorSaveInterval.C:
				e.saveCursor()
			default:
				var (
					events []*models.Event
					err    error
				)

				if !e.resume {
					events, err = e.eventMgr.ListEvents(e.ctx,
						&q.Query{Keywords: q.KeyWords{
							corecommon.StartID: e.cursor.Position,
							corecommon.Limit:   int(batchEventsCount)},
						})
					if err != nil {
						log.Errorf(e.ctx, "failed to list event by offset: %d, limit: %d",
							e.cursor.Position, batchEventsCount)
						time.Sleep(idleWaitInterval)
						continue
					}
					if len(events) == 0 {
						time.Sleep(idleWaitInterval)
						continue
					}
				} else {
					// resume: continue to process the events that are halfway before restart
					lastProcessingCursor, err := e.getLastProcessingCursor()
					if err != nil {
						log.Error(e.ctx, "failed to get last processing cursor")
						time.Sleep(idleWaitInterval)
						continue
					}
					if lastProcessingCursor <= e.cursor.Position {
						e.resume = false
						continue
					}
					events, err = e.eventMgr.ListEventsByRange(e.ctx, e.cursor.Position, lastProcessingCursor)
					if err != nil {
						log.Errorf(e.ctx, "failed to list event by limit %d offset %d", e.cursor.Position, batchEventsCount)
						time.Sleep(idleWaitInterval)
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

func (e *eventHandlerService) getCursor(ctx context.Context) (*models.EventCursor, error) {
	return e.eventMgr.GetCursor(ctx, models.CursorHorizon)
}

// getLastProcessingCursor get the last processing event cursor to resume
func (e *eventHandlerService) getLastProcessingCursor() (uint, error) {
	return e.webhookMgr.GetMaxEventIDOfLog(e.ctx)
}

// saveCursor saves the event id as position in case of resume
func (e *eventHandlerService) saveCursor() {
	if _, err := e.eventMgr.CreateOrUpdateCursor(e.ctx, &models.EventCursor{
		ID:       e.cursor.ID,
		Position: e.cursor.Position,
	}); err != nil {
		log.Errorf(e.ctx, "failed to save cursor(%d), error: %+v",
			e.cursor.Position, err)
	}
}
