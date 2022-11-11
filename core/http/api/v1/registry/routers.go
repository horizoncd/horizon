package registry

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/registries")

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
