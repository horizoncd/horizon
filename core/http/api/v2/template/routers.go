package template

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group(fmt.Sprintf("/apis/core/v2/groups/:%s/templates", _groupParam))
	routes := route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.ListTemplatesByGroupID,
		},
	}
	route.RegisterRoutes(apiGroup, routes)

	apiGroup = engine.Group("/apis/core/v2/templates")
	routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.List,
		},
	}
	route.RegisterRoutes(apiGroup, routes)
}
