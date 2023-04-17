package metrics

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/metrics")

	routes := route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: gin.WrapH(promhttp.Handler()),
		},
	}
	route.RegisterRoutes(api, routes)
}
