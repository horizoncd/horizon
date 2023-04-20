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

package event

import (
	"context"

	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Controller interface {
	ListSupportEvents(ctx context.Context) map[string]string
}

type controller struct {
	eventMgr eventmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		eventMgr: param.EventManager,
	}
}

func (c *controller) ListSupportEvents(ctx context.Context) map[string]string {
	const op = "event controller: list supported events"
	defer wlog.Start(ctx, op).StopPrint()

	return c.eventMgr.ListSupportEvents()
}
