package template

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1/templates")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListTemplate,
		},
		{
			Method: http.MethodGet,
			// TODO: remove this api
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
			HandlerFunc: api.ListTemplatesByGroupID,
		},
	}
	route.RegisterRoutes(apiGroup, routes)

	apiGroup = engine.Group(fmt.Sprintf("/apis/core/v1/templatereleases/:%s", _releaseParam))
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
