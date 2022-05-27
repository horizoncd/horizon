package oauthapp

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

const (
	_groupIDParam            = "groupID"
	_oauthAppClientIDIDParam = "appID"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")
	r := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/oauthapps", _groupIDParam),
			HandlerFunc: api.CreateOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/groups/:%v/oauthapps", _groupIDParam),
			HandlerFunc: api.ListOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDIDParam),
			HandlerFunc: api.GetOauthApp,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDIDParam),
			HandlerFunc: api.UpdateOauthApp,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDIDParam),
			HandlerFunc: api.DeleteOauthApp,
		},
	}
	route.RegisterRoutes(apiGroup, r)
}
