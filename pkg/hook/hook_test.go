package hook

import (
	"context"
	"testing"

	handlermock "g.hz.netease.com/horizon/mock/pkg/hook/handler"
	hhook "g.hz.netease.com/horizon/pkg/hook/hook"
	"github.com/golang/mock/gomock"
)

func TestHook(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockHandler := handlermock.NewMockEventHandler(mockCtl)

	eventHandlers := make([]EventHandler, 0)
	eventHandlers = append(eventHandlers, mockHandler)

	memHook := InMemHook{
		events:        make(chan *hhook.EventCtx, 10),
		eventHandlers: eventHandlers,
		quit:          make(chan bool),
	}

	ctx := context.TODO()
	event1 := hhook.Event{
		EventType: "event1",
		Event:     nil,
	}
	event2 := hhook.Event{
		EventType: "event2",
		Event:     "abc",
	}
	memHook.Push(ctx, event1)
	memHook.Push(ctx, event2)

	mockHandler.EXPECT().Process(&hhook.EventCtx{
		EventType: event1.EventType,
		Event:     event1.Event,
		Ctx:       ctx,
	}).Times(1)
	mockHandler.EXPECT().Process(&hhook.EventCtx{
		EventType: event2.EventType,
		Event:     event2.Event,
		Ctx:       ctx,
	}).Times(1)
	go memHook.Process()
	memHook.Stop()
	memHook.WaitStop()
}
