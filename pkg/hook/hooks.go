package hook

import (
	"context"
	"reflect"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/util/log"
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
	rid, err := requestid.FromContext(ctx)
	if err != nil {
		ctx = context.Background()
	} else {
		ctx = log.WithContext(context.Background(), rid)
	}

	ctxUser, err := user.FromContext(ctx)
	if err != nil {
		log.Error(ctx, "can not find user in context")
	} else {
		ctx = user.WithContext(ctx, ctxUser)
	}

	newEvent := &hook.EventCtx{
		EventType: event.EventType,
		Event:     event.Event,
		Ctx:       ctx,
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
				log.Errorf(event.Ctx, "handler %s, err = %s", reflect.TypeOf(handlerEntry).Name(), err.Error())
			}
		}
	}
	log.Info(context.TODO(), "process ok")
	h.quit <- true
	log.Info(context.TODO(), "channel closed, ProcessExit")
}
