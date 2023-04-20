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

package hook

import (
	"context"
	"math"
	"reflect"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/pkg/hook/hook"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type EventHandler interface {
	Process(event *hook.EventCtx) error
}

type InMemHook struct {
	events        chan *hook.EventCtx
	eventHandlers []EventHandler
	quit          chan bool
}

func NewInMemHook(channelSize int, handlers ...EventHandler) hook.Hook {
	return &InMemHook{
		events:        make(chan *hook.EventCtx, channelSize),
		eventHandlers: handlers,
		quit:          make(chan bool),
	}
}

func (h *InMemHook) Stop() {
	close(h.events)
	log.Info(context.TODO(), "channel closed")
}

func (h *InMemHook) WaitStop() {
	<-h.quit
}

func (h *InMemHook) Push(ctx context.Context, event hook.Event) {
	var newCtx context.Context
	rid, err := requestid.FromContext(ctx)
	if err != nil {
		log.Warning(ctx, "rid not found in ctx")
		newCtx = context.Background()
	} else {
		newCtx = log.WithContext(context.Background(), rid)
	}

	ctxUser, err := common.UserFromContext(ctx)
	if err != nil {
		log.Warning(ctx, "can not find user in context")
	} else {
		newCtx = common.WithContext(newCtx, ctxUser)
	}

	newEvent := &hook.EventCtx{
		EventType:   event.EventType,
		Event:       event.Event,
		Ctx:         newCtx,
		FailedTimes: 0,
	}
	h.events <- newEvent
	log.Infof(ctx, "pushed event, eventType = %s, event = %+v", event.EventType, event.Event)
}

func (h *InMemHook) Process() {
	for event := range h.events {
		log.Infof(event.Ctx, "received event, eventType = %s, event = %+v", event.EventType, event.Event)
		for _, handlerEntry := range h.eventHandlers {
			err := handlerEntry.Process(event)
			if err != nil {
				time.AfterFunc(h.when(event), func() {
					h.events <- event
				})
				log.Errorf(event.Ctx, "handler %s, err = %s", reflect.TypeOf(handlerEntry).Name(), err.Error())
			} else {
				log.Infof(event.Ctx, "processed event, eventType = %s, event = %+v, handler %s,",
					event.EventType, event.Event, reflect.TypeOf(handlerEntry).Name())
			}
		}
	}
	log.Info(context.TODO(), "process ok")
	h.quit <- true
	log.Info(context.TODO(), "channel closed, ProcessExit")
}

func (h *InMemHook) when(event *hook.EventCtx) time.Duration {
	event.FailedTimes++

	backoff := float64(hook.DefaultDelay.Nanoseconds()) * math.Pow(2, float64(event.FailedTimes))
	if backoff > math.MaxInt64 {
		return hook.MaxDelay
	}

	calculated := time.Duration(backoff)
	if calculated > hook.MaxDelay {
		return hook.MaxDelay
	}

	return calculated
}
