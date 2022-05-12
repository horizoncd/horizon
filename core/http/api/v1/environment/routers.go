package environment

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/environments")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListEnvironments,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _environmentIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/regions", _environmentParam),
			HandlerFunc: api.ListEnvironmentRegions,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
