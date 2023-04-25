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
	"time"
)

type EventType string

const (
	CreateApplication EventType = "CreateApplication"
	DeleteApplication EventType = "DeleteApplication"
	CreateCluster     EventType = "CreateCluster"
	DeleteCluster     EventType = "DeleteCluster"
)

var (
	DefaultDelay = 1 * time.Second
	MaxDelay     = 1000 * time.Second
)

type Event struct {
	EventType EventType
	Event     interface{}
}

type EventCtx struct {
	EventType   EventType
	Event       interface{}
	Ctx         context.Context
	FailedTimes uint
}
