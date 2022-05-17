package harbor

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/harbors")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.listAll,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
