package template

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
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
			Pattern:     fmt.Sprintf("/:%v/releases/:%v/schema", _templateParam, _releaseParam),
			HandlerFunc: api.GetTemplateSchema,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: api.GetTemplate,
			Pattern:     fmt.Sprintf("/:%s", _templateParam),
		},
		{
			Method:      http.MethodPut,
			HandlerFunc: api.UpdateTemplate,
			Pattern:     fmt.Sprintf("/:%s", _templateParam),
		},
		{
			Method:      http.MethodDelete,
			HandlerFunc: api.DeleteTemplate,
			Pattern:     fmt.Sprintf("/:%s", _templateParam),
		},
		{
			Method:      http.MethodPost,
			HandlerFunc: api.CreateRelease,
			Pattern:     fmt.Sprintf("/:%s/releases", _templateParam),
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: api.GetReleases,
			Pattern:     fmt.Sprintf("/:%s/releases", _templateParam),
		},
	}
	route.RegisterRoutes(apiGroup, routes)

	apiGroup = engine.Group(fmt.Sprintf("/apis/core/v1/groups/:%s/templates", _groupParam))
	routes = route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: api.CreateTemplate,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: api.GetTemplates,
		},
	}
	route.RegisterRoutes(apiGroup, routes)

	apiGroup = engine.Group(fmt.Sprintf("/apis/core/v1/releases/:%s", _releaseParam))
	routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/schema",
			HandlerFunc: api.GetTemplateSchema,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: api.GetRelease,
		},
		{
			Method:      http.MethodPut,
			HandlerFunc: api.UpdateRelease,
		},
		{
			Method:      http.MethodDelete,
			HandlerFunc: api.DeleteRelease,
		},
		{
			Method:      http.MethodPost,
			HandlerFunc: api.SyncReleaseToRepo,
			Pattern:     "/sync",
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
