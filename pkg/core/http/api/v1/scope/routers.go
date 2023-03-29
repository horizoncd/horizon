package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/scopes",
			HandlerFunc: api.ListScopes,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
