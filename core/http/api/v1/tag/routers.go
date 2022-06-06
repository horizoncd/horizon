package tag

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/tags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/:%v/tags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.Update,
		},
	}
	route.RegisterRoutes(group, routes)
}
