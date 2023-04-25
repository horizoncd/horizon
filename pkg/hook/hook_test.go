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
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	handlermock "github.com/horizoncd/horizon/mock/pkg/hook/handler"
	hhook "github.com/horizoncd/horizon/pkg/hook/hook"
	"github.com/horizoncd/horizon/pkg/util/log"
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
