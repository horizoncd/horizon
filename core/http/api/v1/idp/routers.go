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

package idp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1/idps")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListIDPs,
		},
		{
			Method:      http.MethodPost,
			Pattern:     "/discovery",
			HandlerFunc: api.GetDiscovery,
		},
		{
			Pattern:     "/endpoints",
			Method:      http.MethodGet,
			HandlerFunc: api.ListAuthEndpoints,
		},
		{
			Method:      http.MethodPost,
			HandlerFunc: api.CreateIDP,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodGet,
			HandlerFunc: api.GetByID,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodDelete,
			HandlerFunc: api.DeleteIDP,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodPut,
			HandlerFunc: api.UpdateIDP,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
	engine.GET("/apis/core/v1/login/callback", api.LoginCallback)
	engine.POST("/apis/core/v1/logout", api.Logout)
}
