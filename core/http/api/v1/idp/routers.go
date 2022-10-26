package idp

import (
	"fmt"
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
			Method:      http.MethodPost,
			Pattern:     "/discovery",
			HandlerFunc: api.GetDiscovery,
		},
		{
			Pattern:     "/endpoints",
			Method:      http.MethodGet,
			HandlerFunc: api.ListAuthEndpoints,
		},
		{
			Method:      http.MethodPost,
			HandlerFunc: api.CreateIDP,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodGet,
			HandlerFunc: api.GetByID,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodDelete,
			HandlerFunc: api.DeleteIDP,
		},
		{
			Pattern:     fmt.Sprintf("/:%s", _idp),
			Method:      http.MethodPut,
			HandlerFunc: api.UpdateIDP,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
	engine.GET("/apis/core/v1/login/callback", api.LoginCallback)
	engine.POST("/apis/core/v1/logout", api.Logout)
}
