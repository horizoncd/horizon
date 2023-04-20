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
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (a *API) RegisterRoute(engine *gin.Engine) {
	coreAPI := engine.Group("/apis/core/v2")
	var coreRoutes = route.Routes{
		{
			Pattern:     "/supportevents",
			Method:      http.MethodGet,
			HandlerFunc: a.ListSupportEvents,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
}
