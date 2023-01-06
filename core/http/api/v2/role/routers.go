package role

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v2")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/roles",
			HandlerFunc: api.ListRole,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
