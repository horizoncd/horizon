package user

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v2/users")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", _userIDParam),
			HandlerFunc: api.GetByID,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/self",
			HandlerFunc: api.GetSelf,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", _userIDParam),
			HandlerFunc: api.Update,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/links", _userIDParam),
			HandlerFunc: api.GetLinksByUser,
		},
		{
			Method:      http.MethodPost,
			Pattern:     "/login",
			HandlerFunc: api.LoginWithPassword,
		},
	}
	route.RegisterRoutes(coreGroup, coreRoutes)

	linkGroup := engine.Group("/apis/core/v2/links")
	var linkRoutes = route.Routes{
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _linkIDParam),
			HandlerFunc: api.DeleteLink,
		},
	}
	route.RegisterRoutes(linkGroup, linkRoutes)
}
