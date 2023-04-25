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

package registry

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1/registries")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListAll,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.GetByID,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.UpdateByID,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.DeleteByID,
		}, {
			Method:      http.MethodGet,
			Pattern:     "/kinds",
			HandlerFunc: api.GetKinds,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
