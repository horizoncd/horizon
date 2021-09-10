package metrics

import (
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/metrics")

	var routes = route.Routes{
		{
			"Metrics",
			http.MethodGet,
			"",
			gin.WrapH(promhttp.Handler()),
		},
	}
	route.RegisterRoutes(api, routes)
}
