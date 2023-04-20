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
func (a *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/log", _pipelinerunIDParam),
			HandlerFunc: a.Log,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/stop", _pipelinerunIDParam),
			HandlerFunc: a.Stop,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/diffs", _pipelinerunIDParam),
			HandlerFunc: a.GetDiff,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v", _pipelinerunIDParam),
			HandlerFunc: a.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/pipelineruns", _clusterIDParam),
			HandlerFunc: a.List,
		},
	}

	// get cluster latest pipelinerun log
	// stop pipelinerun for cluster
	// only used for overmind
	frontGroup := engine.Group("/apis/front/v1")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/log", _clusterParam),
			HandlerFunc: a.LatestLogForCluster,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/stop", _clusterParam),
			HandlerFunc: a.StopPipelinerunForCluster,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
