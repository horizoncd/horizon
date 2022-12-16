package application

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiV2Group := engine.Group("/apis/core/v2")
	apiV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/groups/:%v/applications", common.ParamGroupID),
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Get,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/applications/:%v", common.ParamApplicationID),
			HandlerFunc: api.Update,
		},
	}
	route.RegisterRoutes(apiV2Group, apiV2Routes)
}
