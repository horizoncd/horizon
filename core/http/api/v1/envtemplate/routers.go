package envtemplate

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
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/envtemplates/:%v", _applicationIDParam, _envParam),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/envtemplates/:%v", _applicationIDParam, _envParam),
			HandlerFunc: api.Update,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
