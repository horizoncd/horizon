package oauthserver

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

const (
	BasicPath       = "/oauth"
	AuthorizePath   = "/authorize"
	AccessTokenPath = "/access_token"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group(BasicPath)

	var routes = route.Routes{
		{
			Pattern:     AuthorizePath,
			Method:      http.MethodGet,
			HandlerFunc: api.HandleAuthorizationGetReq,
		}, {
			Pattern:     AuthorizePath,
			Method:      http.MethodPost,
			HandlerFunc: api.HandleAuthorizationReq,
		}, {
			Pattern:     AccessTokenPath,
			Method:      http.MethodPost,
			HandlerFunc: api.HandleAccessTokenReq,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
