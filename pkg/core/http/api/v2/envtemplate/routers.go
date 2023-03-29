package envtemplate

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	apiGroup := engine.Group("/apis/core/v2")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/envtemplates", _applicationIDParam),
			HandlerFunc: api.Get,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/envtemplates", _applicationIDParam),
			HandlerFunc: api.Update,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
