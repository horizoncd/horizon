package cluster

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiV2Group := engine.Group("/apis/core/v2")
	apiV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", common.ParamApplicationID),
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/clusters/:%v", common.ParamClusterID),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v", common.ParamClusterID),
			HandlerFunc: api.Get,
		},
	}

	internalV2Group := engine.Group("/apis/internal/v2/clusters")
	internalV2Routes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/deploy/:%v", common.ParamClusterID, common.ParamPipelinerunID),
			HandlerFunc: api.InternalDeploy,
		},
	}

	route.RegisterRoutes(apiV2Group, apiV2Routes)
	route.RegisterRoutes(internalV2Group, internalV2Routes)
}
