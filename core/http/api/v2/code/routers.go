package code

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/front/v2")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "code/listbranch",
			HandlerFunc: api.ListBranch,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "code/listtag",
			HandlerFunc: api.ListTag,
		},
	}
	route.RegisterRoutes(group, routes)
}
