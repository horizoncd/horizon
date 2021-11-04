package hook

import (
	"context"
	"reflect"

	"g.hz.netease.com/horizon/pkg/hook/handler"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/util/log"
)

type Event struct {
	hook.Event
	Ctx context.Context
}

type InMemHook struct {
	events        chan *Event
	eventHandlers []handler.EventHandler
}

func NewInMemHook(channelSize int) hook.Hook {
	return &InMemHook{events: make(chan *Event, channelSize)}
}

func (h *InMemHook) Push(ctx context.Context, event hook.Event) {
	newEvent := &Event{
		Event: event,
		Ctx:   ctx,
	}
	h.events <- newEvent
	log.Info(ctx, "received event %s, event = %+v", event.EventType, event.Event)
}

func (h *InMemHook) Process() {
	for event := range h.events {
		log.Info(event.Ctx, "received event %s, event = %+v", event.EventType, event.Event)
		for _, handlerEntry := range h.eventHandlers {
			err := handlerEntry.Process(event)
			if err != nil {
				log.Info(event.Ctx, "handler ", reflect.TypeOf(handlerEntry))
			}
		}
	}
}
