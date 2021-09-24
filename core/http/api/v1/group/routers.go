package group

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, c *Controller) {
	coreAPI := engine.Group("/apis/core/v1/groups")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: c.CreateGroup,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: c.DeleteGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: c.GetGroup,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: c.GetGroupByPath,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", ParamGroupID),
			HandlerFunc: c.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/groups", ParamGroupID),
			HandlerFunc: c.GetSubGroups,
		},
	}

	frontAPI := engine.Group("/apis/front/v1/groups")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/children", ParamGroupID),
			HandlerFunc: c.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/search",
			HandlerFunc: c.SearchGroups,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/transfer", ParamGroupID),
			HandlerFunc: c.TransferGroup,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
