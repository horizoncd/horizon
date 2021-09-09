package group

import (
	"net/http"

	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, c *Controller) {
	api := engine.Group("/api/v1/groups")

	var routes = route.Routes{
		{
			"CreateGroup",
			http.MethodPost,
			"",
			c.CreateGroup,
		},
		{
			"DeleteGroup",
			http.MethodDelete,
			"/:groupId",
			c.DeleteGroup,
		},
		{
			"GetGroup",
			http.MethodGet,
			"/:groupId",
			c.GetGroup,
		},
		{
			"GetGroupByPath",
			http.MethodGet,
			"",
			c.GetGroupByPath,
		},
		{
			"UpdateGroup",
			http.MethodPut,
			"/:groupId",
			c.UpdateGroup,
		},
		{
			"GetChildren",
			http.MethodGet,
			"/:groupId/children",
			c.GetChildren,
		},
		{
			"GetSubGroups",
			http.MethodGet,
			"/:groupId/subgroups",
			c.GetSubGroups,
		},
		{
			"SearchGroups",
			http.MethodGet,
			"/search",
			c.SearchGroups,
		},
	}
	route.RegisterRoutes(api, routes)
}
