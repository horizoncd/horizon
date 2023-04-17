package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes.
func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/health")

	routes := route.Routes{
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
