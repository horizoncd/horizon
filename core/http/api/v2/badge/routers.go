package badge

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (a *API) RegisterRoute(engine *gin.Engine) {
	apiV2Group := engine.Group("/apis/core/v2")
	apiV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/:%v/badges", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: a.Create,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/badges", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: a.List,
		},
		{
			Method: http.MethodGet,
			Pattern: fmt.Sprintf("/:%v/:%v/badges/:%v",
				common.ParamResourceType, common.ParamResourceID, _paramBadgeIDorName),
			HandlerFunc: a.Get,
		},
		{
			Method: http.MethodPut,
			Pattern: fmt.Sprintf("/:%v/:%v/badges/:%v",
				common.ParamResourceType, common.ParamResourceID, _paramBadgeIDorName),
			HandlerFunc: a.Update,
		},
		{
			Method: http.MethodDelete,
			Pattern: fmt.Sprintf("/:%v/:%v/badges/:%v",
				common.ParamResourceType, common.ParamResourceID, _paramBadgeIDorName),
			HandlerFunc: a.Delete,
		},
	}
	route.RegisterRoutes(apiV2Group, apiV2Routes)
}
