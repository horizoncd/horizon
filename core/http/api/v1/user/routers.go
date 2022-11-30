package user

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	coreGroup := engine.Group("/apis/core/v1/users")
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

	linkGroup := engine.Group("/apis/core/v1/links")
	var linkRoutes = route.Routes{
		{
			Method: http.MethodDelete,
			// Deprecated: /apis/front/v1/users/search is not recommend, use /apis/core/v1/users instead
			Pattern:     fmt.Sprintf("/:%v", _linkIDParam),
			HandlerFunc: api.DeleteLink,
		},
	}
	route.RegisterRoutes(linkGroup, linkRoutes)

	frontGroup := engine.Group("/apis/front/v1/users")
	var frontRoutes = route.Routes{
		{
			Method: http.MethodGet,
			// Deprecated: /apis/front/v1/users/search is not recommend, use /apis/core/v1/users instead
			Pattern:     "/search",
			HandlerFunc: api.List,
		},
	}
	route.RegisterRoutes(frontGroup, frontRoutes)
}
