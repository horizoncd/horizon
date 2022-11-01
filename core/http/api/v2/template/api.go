package template

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
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
}
