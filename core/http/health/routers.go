package health

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/health")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: healthCheck,
		},
	}
	route.RegisterRoutes(api, routes)
}

func healthCheck(c *gin.Context) {
	response.Success(c)
}
