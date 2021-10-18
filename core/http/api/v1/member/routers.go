package member

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/groups/:%v/members", _paramGroupID),
			HandlerFunc: api.ListGroupMember,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/members", _paramGroupID),
			HandlerFunc: api.CreateMember,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/members/:%v", _paramMemberID),
			HandlerFunc: api.UpdateMember,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/members/:%v", _paramMemberID),
			HandlerFunc: api.DeleteMember,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
