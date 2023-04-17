package region

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes.
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v1/regions")

	routes := route.Routes{
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
