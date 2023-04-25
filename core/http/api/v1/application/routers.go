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

package application

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/applications", common.ParamGroupID),
			HandlerFunc: api.Create,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/applications",
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/selectableregions", common.ParamApplicationID),
			HandlerFunc: api.GetSelectableRegionsByEnv,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Update,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Delete,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/applications/:%v/transfer", common.ParamApplicationID),
			HandlerFunc: api.Transfer,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/pipelinestats", common.ParamApplicationID),
			HandlerFunc: api.GetApplicationPipelineStats,
		},
	}

	route.RegisterRoutes(apiGroup, routes)

	frontGroup := engine.Group("/apis/front/v1/applications")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/searchmyapplications",
			HandlerFunc: api.ListSelf,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchapplications",
			HandlerFunc: api.List,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
}
