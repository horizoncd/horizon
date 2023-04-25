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

package webhook

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/pkg/server/route"
)

const (
	_resourceTypeParam = "resourceType"
	_resourceIDParam   = "resourceID"
	_webhookIDParam    = "webhookID"
	_webhookLogIDParam = "webhookLogID"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	group := engine.Group("/apis/core/v2")
	var routers = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/:%v/webhooks", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.CreateWebhook,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/webhooks", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.ListWebhooks,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/webhooks/:%v", _webhookIDParam),
			HandlerFunc: api.UpdateWebhook,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/webhooks/:%v", _webhookIDParam),
			HandlerFunc: api.GetWebhook,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/webhooks/:%v", _webhookIDParam),
			HandlerFunc: api.DeleteWebhook,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/webhooks/:%v/logs", _webhookIDParam),
			HandlerFunc: api.ListWebhookLogs,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/webhooklogs/:%v", _webhookLogIDParam),
			HandlerFunc: api.GetWebhookLog,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/webhooklogs/:%v/resend", _webhookLogIDParam),
			HandlerFunc: api.ResendWebhook,
		},
	}
	route.RegisterRoutes(group, routers)
}
