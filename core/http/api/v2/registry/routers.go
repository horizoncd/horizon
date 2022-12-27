package registry

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v2/registries")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListAll,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.GetByID,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.UpdateByID,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _registryIDParam),
			HandlerFunc: api.DeleteByID,
		}, {
			Method:      http.MethodGet,
			Pattern:     "/kinds",
			HandlerFunc: api.GetKinds,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
