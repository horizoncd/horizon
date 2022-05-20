package oauthserver

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/oauth")

	var routes = route.Routes{
		{
			Pattern:     "/authorize",
			Method:      http.MethodGet,
			HandlerFunc: api.HandleAuthorizationGetReq,
		}, {
			Pattern:     "/authorize",
			Method:      http.MethodPost,
			HandlerFunc: api.HandleAuthorizationReq,
		}, {
			Pattern:     "/access_token",
			Method:      http.MethodPost,
			HandlerFunc: api.HandleAccessTokenReq,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
