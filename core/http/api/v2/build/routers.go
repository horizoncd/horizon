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

package build

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	apiV2Group := engine.Group("/apis/front/v2")
	apiV2Route := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "buildschema",
			HandlerFunc: api.Get,
		},
	}
	route.RegisterRoutes(apiV2Group, apiV2Route)
}
