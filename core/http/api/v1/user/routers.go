package user

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/front/v1/users")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/search",
			HandlerFunc: api.Search,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
