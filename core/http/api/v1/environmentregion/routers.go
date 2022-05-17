package environmentregion

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/environmentregions")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListAll,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/enable", _environmentRegionIDParam),
			HandlerFunc: api.Enable,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/disable", _environmentRegionIDParam),
			HandlerFunc: api.Disable,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/setdefault", _environmentRegionIDParam),
			HandlerFunc: api.SetDefault,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
