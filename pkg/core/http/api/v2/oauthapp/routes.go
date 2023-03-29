package oauthapp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

const (
	_groupIDParam          = "groupID"
	_oauthAppClientIDParam = "appID"
	_oauthClientSecretID   = "secretID"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v2")
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
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.GetOauthApp,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.UpdateOauthApp,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/oauthapps/:%v", _oauthAppClientIDParam),
			HandlerFunc: api.DeleteOauthApp,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/oauthapps/:%v/clientsecret", _oauthAppClientIDParam),
			HandlerFunc: api.ListSecret,
		}, {
			Method: http.MethodDelete,
			Pattern: fmt.Sprintf("/oauthapps/:%v/clientsecret/:%v",
				_oauthAppClientIDParam, _oauthClientSecretID),
			HandlerFunc: api.DeleteSecret,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/oauthapps/:%v/clientsecret", _oauthAppClientIDParam),
			HandlerFunc: api.CreateSecret,
		},
	}
	route.RegisterRoutes(apiGroup, r)
}
