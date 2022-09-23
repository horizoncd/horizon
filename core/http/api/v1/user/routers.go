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
	}
	route.RegisterRoutes(coreGroup, coreRoutes)

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

	loginGroup := engine.Group("/apis/login/v1")
	var apiRoutes = route.Routes{
		{
			Method: http.MethodGet,
			// Deprecated: /apis/login/v1/status is not recommend, use /apis/core/v1/users/self instead
			Pattern:     "/status",
			HandlerFunc: api.GetSelf,
		},
	}

	route.RegisterRoutes(loginGroup, apiRoutes)
}
