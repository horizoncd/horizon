package pipelinerun

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/log", _pipelinerunIDParam),
			HandlerFunc: api.Log,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/stop", _pipelinerunIDParam),
			HandlerFunc: api.Stop,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/diffs", _pipelinerunIDParam),
			HandlerFunc: api.GetDiff,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v", _pipelinerunIDParam),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/pipelineruns", _clusterIDParam),
			HandlerFunc: api.List,
		},
	}

	// get cluster latest pipelinerun log
	// stop pipelinerun for cluster
	// only used for overmind
	frontGroup := engine.Group("/apis/front/v1")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/log", _clusterParam),
			HandlerFunc: api.LatestLogForCluster,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/stop", _clusterParam),
			HandlerFunc: api.StopPipelinerunForCluster,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
