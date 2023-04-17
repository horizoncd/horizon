package role

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes.
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")

	routes := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/roles",
			HandlerFunc: api.ListRole,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
