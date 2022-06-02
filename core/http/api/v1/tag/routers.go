package tag

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/core/v1/tags")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPost,
			HandlerFunc: api.Update,
		},
	}
	route.RegisterRoutes(group, routes)
}
