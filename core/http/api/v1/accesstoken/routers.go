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

package accesstoken

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v1")
	var coreRouters = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/personalaccesstokens",
			HandlerFunc: api.CreatePersonalAccessToken,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.CreateResourceAccessToken,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/personalaccesstokens",
			HandlerFunc: api.ListPersonalAccessTokens,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.ListResourceAccessTokens,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/personalaccesstokens/:%s", common.ParamAccessTokenID),
			HandlerFunc: api.RevokePersonalAccessToken,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/accesstokens/:%s", common.ParamAccessTokenID),
			HandlerFunc: api.RevokeResourceAccessToken,
		},
	}

	route.RegisterRoutes(coreGroup, coreRouters)
}
