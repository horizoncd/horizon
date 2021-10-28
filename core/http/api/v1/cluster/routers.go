package cluster

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
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", _applicationIDParam),
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/clusters/:%v", _clusterIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v", _clusterIDParam),
			HandlerFunc: api.Get,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
