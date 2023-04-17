package member

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes.
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")

	routes := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/groups/:%v/members", _paramGroupID),
			HandlerFunc: api.ListGroupMember,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/members", _paramGroupID),
			HandlerFunc: api.CreateGroupMember,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/members", _paramApplicationID),
			HandlerFunc: api.ListApplicationMember,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/members", _paramApplicationID),
			HandlerFunc: api.CreateApplicationMember,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/members", _paramApplicationClusterID),
			HandlerFunc: api.ListApplicationClusterMember,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/members", _paramApplicationClusterID),
			HandlerFunc: api.CreateApplicationClusterMember,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/templates/:%v/members", _paramTemplateID),
			HandlerFunc: api.ListTemplateMember,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/templates/:%v/members", _paramTemplateID),
			HandlerFunc: api.CreateTemplateMember,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/members/:%v", _paramMemberID),
			HandlerFunc: api.UpdateMember,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/members/:%v", _paramMemberID),
			HandlerFunc: api.DeleteMember,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
