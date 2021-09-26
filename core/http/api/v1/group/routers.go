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
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: a.DeleteGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: a.GetGroup,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: a.GetGroupByPath,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: a.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/groups", ParamGroupID),
			HandlerFunc: a.GetSubGroups,
		},
	}

	frontAPI := engine.Group("/apis/front/v1/groups")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/children", ParamGroupID),
			HandlerFunc: a.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/search",
			HandlerFunc: a.SearchGroups,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/transfer", ParamGroupID),
			HandlerFunc: a.TransferGroup,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
