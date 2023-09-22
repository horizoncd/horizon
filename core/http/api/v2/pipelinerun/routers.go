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

package pipelinerun

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v2")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/log", _pipelinerunIDParam),
			HandlerFunc: api.Log,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/stop", _pipelinerunIDParam),
			HandlerFunc: api.Stop,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/diffs", _pipelinerunIDParam),
			HandlerFunc: api.GetDiff,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v", _pipelinerunIDParam),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/pipelineruns", _clusterIDParam),
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/run", _pipelinerunIDParam),
			HandlerFunc: api.Execute,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/forcerun", _pipelinerunIDParam),
			HandlerFunc: api.ExecuteForce,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/cancel", _pipelinerunIDParam),
			HandlerFunc: api.Cancel,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/checkruns", _pipelinerunIDParam),
			HandlerFunc: api.CreateCheckRun,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/checkruns/:%v", _checkrunIDParam),
			HandlerFunc: api.GetCheckRun,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/checkruns", _pipelinerunIDParam),
			HandlerFunc: api.ListCheckRuns,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/checkruns/:%v", _checkrunIDParam),
			HandlerFunc: api.UpdateCheckRun,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/messages", _pipelinerunIDParam),
			HandlerFunc: api.ListPrMessages,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/messages", _pipelinerunIDParam),
			HandlerFunc: api.CreatePrMessage,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
