package template

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/templates")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListTemplate,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/releases", _templateParam),
			HandlerFunc: api.ListTemplateRelease,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/releases/:%v/schema", _templateParam, _releaseParam),
			HandlerFunc: api.GetTemplateSchema,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
