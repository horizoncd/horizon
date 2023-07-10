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

package oauthserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

const (
	BasicPath       = "/login/oauth"
	AuthorizePath   = "/authorize"
	AccessTokenPath = "/access_token"
)

func (a *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group(BasicPath)

	var routes = route.Routes{
		{
			Pattern:     AuthorizePath,
			Method:      http.MethodGet,
			HandlerFunc: a.HandleAuthorizationGetReq,
		}, {
			Pattern:     AuthorizePath,
			Method:      http.MethodPost,
			HandlerFunc: a.HandleAuthorizationReq,
		}, {
			Pattern:     AccessTokenPath,
			Method:      http.MethodPost,
			HandlerFunc: a.HandleAccessTokenReq,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
