package hook

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	handlermock "github.com/horizoncd/horizon/mock/pkg/hook/handler"
	hhook "github.com/horizoncd/horizon/pkg/hook/hook"
	"github.com/horizoncd/horizon/pkg/log"
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

	requestID := "123"
	ctx := context.WithValue(context.TODO(), requestid.HeaderXRequestID, "123") // nolint
	ctx = log.WithContext(ctx, requestID)
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

	mockHandler.EXPECT().Process(gomock.Any()).Times(1)
	mockHandler.EXPECT().Process(gomock.Any()).Times(1)
	go memHook.Process()
	memHook.Stop()
	memHook.WaitStop()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
