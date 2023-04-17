package pipelinerun

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes.
func (a *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1")
	routes := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/log", _pipelinerunIDParam),
			HandlerFunc: a.Log,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/stop", _pipelinerunIDParam),
			HandlerFunc: a.Stop,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v/diffs", _pipelinerunIDParam),
			HandlerFunc: a.GetDiff,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/pipelineruns/:%v", _pipelinerunIDParam),
			HandlerFunc: a.Get,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/pipelineruns", _clusterIDParam),
			HandlerFunc: a.List,
		},
	}

	// get cluster latest pipelinerun log
	// stop pipelinerun for cluster
	// only used for overmind
	frontGroup := engine.Group("/apis/front/v1")
	frontRoutes := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/log", _clusterParam),
			HandlerFunc: a.LatestLogForCluster,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/stop", _clusterParam),
			HandlerFunc: a.StopPipelinerunForCluster,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
