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

package tag

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	group := engine.Group("/apis/core/v2")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/tags", common.ParamClusterID),
			HandlerFunc: api.ListClusterTags,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/tags", common.ParamApplicationID),
			HandlerFunc: api.ListApplicationTags,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/:%v/tags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/subresourcetags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.ListSubResourceTags,
		}, {
			Method:      http.MethodPost,
			Pattern:     "/metatags",
			HandlerFunc: api.CreateMetatags,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/metatags",
			HandlerFunc: api.GetMetatagsByKey,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/metatagkeys",
			HandlerFunc: api.GetMetatagKeys,
		},
	}
	route.RegisterRoutes(group, routes)
}
