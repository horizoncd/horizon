package group

import (
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, c *Controller) {
	api := engine.Group("/api/v1/groups")

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
