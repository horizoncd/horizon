package application

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/applications", common.ParamGroupID),
			HandlerFunc: api.Create,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/selectableregions", common.ParamApplicationID),
			HandlerFunc: api.GetSelectableRegionsByEnv,
		},
		{
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Update,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Delete,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/applications/:%v/transfer", common.ParamApplicationID),
			HandlerFunc: api.Transfer,
		},
	}

	frontGroup := engine.Group("/apis/front/v1")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/applications/searchapplications",
			HandlerFunc: api.SearchApplication,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/applications/searchmyapplications",
			HandlerFunc: api.SearchMyApplication,
		},
	}

	apiV2Group := engine.Group("/apis/core/v2")
	apiV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/applications", common.ParamGroupID),
			HandlerFunc: api.CreateV2,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(apiV2Group, apiV2Routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
