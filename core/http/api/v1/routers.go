package v1

import (
	"net/http"

	"g.hz.netease.com/horizon/server/middleware/log"
	"g.hz.netease.com/horizon/server/middleware/requestid"
	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	var c, _ = NewController()
	api.Use(requestid.Middleware())
	api.Use(log.Middleware())

	var routes = route.Routes{
		{
			"Test",
			http.MethodGet,
			"/test",
			c.Test,
		},
	}
	route.RegisterRoutes(api, routes)
}

type Controller struct {}

func NewController() (*Controller, error) {

	return &Controller{
	}, nil
}
