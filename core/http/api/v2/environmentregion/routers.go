package environmentregion

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v2/environmentregions")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/setdefault", _environmentRegionIDParam),
			HandlerFunc: api.SetDefault,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _environmentRegionIDParam),
			HandlerFunc: api.DeleteByID,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
