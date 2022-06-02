package harbor

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1/harbors")

	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListAll,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Create,
		}, {
			Method:      http.MethodPut,
			Pattern:     fmt.Sprintf("/:%v", _harborIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/:%v", _harborIDParam),
			HandlerFunc: api.DeleteByID,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
