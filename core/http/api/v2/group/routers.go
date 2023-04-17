package group

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes.
func (a *API) RegisterRoute(engine *gin.Engine) {
	coreAPI := engine.Group("/apis/core/v2/groups")
	coreRoutes := route.Routes{
		{
			Method:      http.MethodPost,
			HandlerFunc: a.CreateGroup,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%s/groups", _paramGroupID),
			HandlerFunc: a.CreateSubGroup,
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
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s", _paramGroupID),
			HandlerFunc: a.UpdateGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/groups", _paramGroupID),
			HandlerFunc: a.GetSubGroups,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/transfer", _paramGroupID),
			HandlerFunc: a.TransferGroup,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%s/regionselectors", _paramGroupID),
			HandlerFunc: a.UpdateRegionSelector,
		},
	}

	frontAPI := engine.Group("/apis/front/v2/groups")
	frontRoutes := route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: a.GetGroupByFullPath,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/authedgroups",
			HandlerFunc: a.ListAuthedGroup,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/children", _paramGroupID),
			HandlerFunc: a.GetChildren,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchgroups",
			HandlerFunc: a.SearchGroups,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchchildren",
			HandlerFunc: a.SearchChildren,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
	route.RegisterRoutes(frontAPI, frontRoutes)
}
