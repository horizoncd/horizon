package application

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		// TODO(gjq): remove Create/Update/Delete application router, after migration, add these routers back.
		// {
		// 	Method:      http.MethodPost,
		// 	Pattern:     fmt.Sprintf("/groups/:%v/applications", _groupIDParam),
		// 	HandlerFunc: api.Create,
		// },
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v", _applicationIDParam),
			HandlerFunc: api.Get,
		},
		// {
		// 	Method:      http.MethodPut,
		// 	Pattern:     fmt.Sprintf("/applications/:%v", _applicationIDParam),
		// 	HandlerFunc: api.Update,
		// },
		// {
		// 	Method:      http.MethodDelete,
		// 	Pattern:     fmt.Sprintf("/applications/:%v", _applicationIDParam),
		// 	HandlerFunc: api.Delete,
		// },
	}

	frontGroup := engine.Group("/apis/front/v1")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/applications/searchapplications",
			HandlerFunc: api.SearchApplication,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
