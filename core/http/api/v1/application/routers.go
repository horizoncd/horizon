package application

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoutes(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/applications", common.ParamGroupID),
			HandlerFunc: api.Create,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/applications",
			HandlerFunc: api.List,
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
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/pipelinestats", common.ParamApplicationID),
			HandlerFunc: api.GetApplicationPipelineStats,
		},
	}

	route.RegisterRoutes(apiGroup, routes)

	frontGroup := engine.Group("/apis/front/v1/applications")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/searchmyapplications",
			HandlerFunc: api.ListSelf,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/searchapplications",
			HandlerFunc: api.List,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
}
