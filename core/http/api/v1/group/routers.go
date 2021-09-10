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
			http.MethodGet,
			"/search",
			c.SearchGroups,
		},
		{
			http.MethodPost,
			"",
			c.CreateGroup,
		},
		{
			http.MethodDelete,
			"/:groupId",
			c.DeleteGroup,
		},
		{
			http.MethodGet,
			"/:groupId",
			c.GetGroup,
		},
		{
			http.MethodGet,
			"",
			c.GetGroupByPath,
		},
		{
			http.MethodPut,
			"/:groupId",
			c.UpdateGroup,
		},
		{
			http.MethodGet,
			"/:groupId/children",
			c.GetChildren,
		},
		{
			http.MethodGet,
			"/:groupId/subgroups",
			c.GetSubGroups,
		},
	}
	route.RegisterRoutes(api, routes)
}
