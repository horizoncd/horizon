package group

import (
	"g.hz.netease.com/horizon/server/middleware/log"
	"g.hz.netease.com/horizon/server/middleware/requestid"
	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, c *Controller) {
	api := engine.Group("/api/v1/groups")
	api.Use(requestid.Middleware())
	api.Use(log.Middleware())

	var routes = route.Routes{
		{
			"CreateGroup",
			http.MethodPost,
			"",
			c.CreateGroup,
		},
	}
	route.RegisterRoutes(api, routes)
}
