package cluster

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
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", _applicationIDParam),
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/clusters", _applicationIDParam),
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/clusters/:%v", _clusterIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v", _clusterIDParam),
			HandlerFunc: api.Get,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/clusters/:%v", _clusterIDParam),
			HandlerFunc: api.Delete,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/builddeploy", _clusterIDParam),
			HandlerFunc: api.BuildDeploy,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/diffs", _clusterIDParam),
			HandlerFunc: api.GetDiff,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/status", _clusterIDParam),
			HandlerFunc: api.ClusterStatus,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/restart", _clusterIDParam),
			HandlerFunc: api.Restart,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/deploy", _clusterIDParam),
			HandlerFunc: api.Deploy,
		},
	}

	frontGroup := engine.Group("/apis/front/v1/clusters")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/searchclusters",
			HandlerFunc: api.ListByNameFuzzily,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _clusterParam),
			HandlerFunc: api.GetByName,
		},
	}

	internalGroup := engine.Group("/apis/internal/v1/clusters")
	var internalRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/deploy", _clusterIDParam),
			HandlerFunc: api.InternalDeploy,
		},
	}
	// TODO use middleware to auth token
	internalGroup.Use()

	route.RegisterRoutes(apiGroup, routes)
	route.RegisterRoutes(frontGroup, frontRoutes)
	route.RegisterRoutes(internalGroup, internalRoutes)
}
