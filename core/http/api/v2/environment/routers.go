package environment

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes.
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v2/environments")
	routes := route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListEnvironments,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _environmentParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v", _environmentParam),
			HandlerFunc: api.GetByID,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _environmentParam),
			HandlerFunc: api.Delete,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
