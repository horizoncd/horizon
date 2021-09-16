package group

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, c *Controller) {
	restAPI := engine.Group("/api/rest/v1/groups")
	var routes = route.Routes{
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
			Pattern:     fmt.Sprintf("/:%s/subgroups", ParamGroupID),
			HandlerFunc: c.GetSubGroups,
		},
	}

	frontAPI := engine.Group("/api/front/v1/groups")
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
	}

	route.RegisterRoutes(restAPI, routes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
