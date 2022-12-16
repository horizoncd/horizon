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

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/core/v1")
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
