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

package group

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (a *API) RegisterRoute(engine *gin.Engine) {
	coreAPI := engine.Group("/apis/core/v1/groups")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: a.CreateGroup,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%s/groups", _paramGroupID),
			HandlerFunc: a.CreateSubGroup,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.DeleteGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.GetGroup,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/groups", _paramGroupID),
			HandlerFunc: a.GetSubGroups,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/transfer", _paramGroupID),
			HandlerFunc: a.TransferGroup,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/regionselectors", _paramGroupID),
			HandlerFunc: a.UpdateRegionSelector,
		},
	}

	frontAPI := engine.Group("/apis/front/v1/groups")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: a.GetGroupByFullPath,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/authedgroups",
			HandlerFunc: a.ListAuthedGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/children", _paramGroupID),
			HandlerFunc: a.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchgroups",
			HandlerFunc: a.SearchGroups,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchchildren",
			HandlerFunc: a.SearchChildren,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
