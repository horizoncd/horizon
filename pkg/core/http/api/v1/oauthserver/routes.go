package oauthserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

const (
	BasicPath       = "/login/oauth"
	AuthorizePath   = "/authorize"
	AccessTokenPath = "/access_token"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
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
