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
	"github.com/horizoncd/horizon/core/controller/event"
	"github.com/horizoncd/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

type API struct {
	eventCtl event.Controller
}

// NewAPI initializes a new event api
func NewAPI(controller event.Controller) *API {
	return &API{
		eventCtl: controller,
	}
}

// ListSupportEvents list actions categorized based on resources
func (a *API) ListSupportEvents(c *gin.Context) {
	response.SuccessWithData(c, a.eventCtl.ListSupportEvents(c))
}
