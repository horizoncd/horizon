package idp

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/idps")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListIDPs,
		},
		{
			Pattern:     "/endpoints",
			Method:      http.MethodGet,
			HandlerFunc: api.ListAuthEndpoints,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
	engine.GET("/apis/core/v1/login/callback", api.LoginCallback)
	engine.POST("/apis/core/v1/logout", api.Logout)
}
