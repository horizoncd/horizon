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

package oauthapp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

const (
	_groupIDParam          = "groupID"
	_oauthAppClientIDParam = "appID"
	_oauthClientSecretID   = "secretID"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")
	r := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/oauthapps", _groupIDParam),
			HandlerFunc: api.CreateOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/groups/:%v/oauthapps", _groupIDParam),
			HandlerFunc: api.ListOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.GetOauthApp,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.UpdateOauthApp,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.DeleteOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/oauthapps/:%v/clientsecret", _oauthAppClientIDParam),
			HandlerFunc: api.ListSecret,
		}, {
			Method: http.MethodDelete,
			Pattern: fmt.Sprintf("/oauthapps/:%v/clientsecret/:%v",
				_oauthAppClientIDParam, _oauthClientSecretID),
			HandlerFunc: api.DeleteSecret,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/oauthapps/:%v/clientsecret", _oauthAppClientIDParam),
			HandlerFunc: api.CreateSecret,
		},
	}
	route.RegisterRoutes(apiGroup, r)
}
