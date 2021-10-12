package user

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	frontGroup := engine.Group("/apis/front/v1/users")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/search",
			HandlerFunc: api.Search,
		},
	}

	apiGroup := engine.Group("/apis/login/v1")
	var apiRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/status",
			HandlerFunc: api.Status,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
	route.RegisterRoutes(apiGroup, apiRoutes)
}
