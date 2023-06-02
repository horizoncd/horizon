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

package cluster

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	apiV2Group := engine.Group("/apis/core/v2")
	apiV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", common.ParamApplicationID),
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", common.ParamApplicationID),
			HandlerFunc: api.ListByApplication,
		}, {
			Method:      http.MethodGet,
			Pattern:     "/clusters",
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/clusters/:%v", common.ParamClusterID),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v", common.ParamClusterID),
			HandlerFunc: api.Get,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/clusters/:%v", common.ParamClusterID),
			HandlerFunc: api.Delete,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/builddeploy", common.ParamClusterID),
			HandlerFunc: api.BuildDeploy,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/diffs", common.ParamClusterID),
			HandlerFunc: api.GetDiff,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/step", common.ParamClusterID),
			HandlerFunc: api.GetStep,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/resourcetree", common.ParamClusterID),
			HandlerFunc: api.GetResourceTree,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/status", common.ParamClusterID),
			HandlerFunc: api.ClusterStatus,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/buildstatus", common.ParamClusterID),
			HandlerFunc: api.ClusterPipelinerunStatus,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/restart", common.ParamClusterID),
			HandlerFunc: api.Restart,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/deploy", common.ParamClusterID),
			HandlerFunc: api.Deploy,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/rollback", common.ParamClusterID),
			HandlerFunc: api.Rollback,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/action", common.ParamClusterID),
			HandlerFunc: api.ExecuteAction,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/containerlog", common.ParamClusterID),
			HandlerFunc: api.GetContainerLog,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/exec", common.ParamClusterID),
			HandlerFunc: api.Exec,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/dashboards", common.ParamClusterID),
			HandlerFunc: api.GetGrafanaDashBoard,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/pod", common.ParamClusterID),
			HandlerFunc: api.GetClusterPod,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/clusters/:%v/pods", common.ParamClusterID),
			HandlerFunc: api.DeleteClusterPods,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/free", common.ParamClusterID),
			HandlerFunc: api.Free,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/events", common.ParamClusterID),
			HandlerFunc: api.PodEvents,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/outputs", common.ParamClusterID),
			HandlerFunc: api.GetOutput,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/containers", common.ParamClusterID),
			HandlerFunc: api.GetContainers,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/favorite", common.ParamClusterID),
			HandlerFunc: api.AddFavorite,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/clusters/:%v/favorite", common.ParamClusterID),
			HandlerFunc: api.DeleteFavorite,
		},
	}

	frontV2Group := engine.Group("/apis/front/v2/clusters")
	var frontV2Routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", common.ParamClusterName),
			HandlerFunc: api.GetByName,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchclusters",
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchmyclusters",
			HandlerFunc: api.ListSelf,
		},
	}

	internalV2Group := engine.Group("/apis/internal/v2/clusters")
	internalV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/deploy", common.ParamClusterID),
			HandlerFunc: api.InternalDeploy,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/status", common.ParamClusterID),
			HandlerFunc: api.InternalClusterStatus,
		},
	}

	route.RegisterRoutes(apiV2Group, apiV2Routes)
	route.RegisterRoutes(frontV2Group, frontV2Routes)
	route.RegisterRoutes(internalV2Group, internalV2Routes)
}
