package application

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/applications")
	var routes = route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _applicationParam),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _applicationParam),
			HandlerFunc: api.Update,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
