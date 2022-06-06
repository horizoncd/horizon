package region

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/kubernetes")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListRegions,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _regionIDParam),
			HandlerFunc: api.GetByID,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/tags", _regionIDParam),
			HandlerFunc: api.ListRegionTags,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _regionIDParam),
			HandlerFunc: api.UpdateByID,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _regionIDParam),
			HandlerFunc: api.DeleteByID,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
