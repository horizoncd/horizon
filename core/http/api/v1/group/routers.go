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
			Method:      http.MethodGet,
			Pattern:     "/search",
			HandlerFunc: c.SearchGroups,
		},
		{
			Method:      http.MethodPost,
			HandlerFunc: c.CreateGroup,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     "/:groupId",
			HandlerFunc: c.DeleteGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/:groupId",
			HandlerFunc: c.GetGroup,
		},
		{
			Method:      http.MethodGet,
			HandlerFunc: c.GetGroupByPath,
		},
		{
			Method:      http.MethodPut,
			Pattern:     "/:groupId",
			HandlerFunc: c.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/:groupId/children",
			HandlerFunc: c.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/:groupId/subgroups",
			HandlerFunc: c.GetSubGroups,
		},
	}
	route.RegisterRoutes(api, routes)
}
