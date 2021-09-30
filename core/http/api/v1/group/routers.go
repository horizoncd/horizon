package group

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, a *API) {
	coreAPI := engine.Group("/apis/core/v1/groups")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: a.CreateGroup,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.DeleteGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.GetGroup,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: a.GetGroupByFullPath,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/groups", _paramGroupID),
			HandlerFunc: a.GetSubGroups,
		},
	}

	frontAPI := engine.Group("/apis/front/v1/groups")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/children", _paramGroupID),
			HandlerFunc: a.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/search-groups",
			HandlerFunc: a.SearchGroups,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/search-children",
			HandlerFunc: a.SearchChildren,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/transfer", _paramGroupID),
			HandlerFunc: a.TransferGroup,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
