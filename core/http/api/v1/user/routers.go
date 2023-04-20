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

package user

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v1/users")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", _userIDParam),
			HandlerFunc: api.GetByID,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/self",
			HandlerFunc: api.GetSelf,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", _userIDParam),
			HandlerFunc: api.Update,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/links", _userIDParam),
			HandlerFunc: api.GetLinksByUser,
		},
		{
			Method:      http.MethodPost,
			Pattern:     "/login",
			HandlerFunc: api.LoginWithPassword,
		},
	}
	route.RegisterRoutes(coreGroup, coreRoutes)

	linkGroup := engine.Group("/apis/core/v1/links")
	var linkRoutes = route.Routes{
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _linkIDParam),
			HandlerFunc: api.DeleteLink,
		},
	}
	route.RegisterRoutes(linkGroup, linkRoutes)

	frontGroup := engine.Group("/apis/front/v1/users")
	var frontRoutes = route.Routes{
		{
			Method: http.MethodGet,
			// Deprecated: /apis/front/v1/users/search is not recommend, use /apis/core/v1/users instead
			Pattern:     "/search",
			HandlerFunc: api.List,
		},
	}
	route.RegisterRoutes(frontGroup, frontRoutes)
}
