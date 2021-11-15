package code

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/front/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "code/listbranch",
			HandlerFunc: api.ListBranch,
		},
	}
	route.RegisterRoutes(group, routes)
}
