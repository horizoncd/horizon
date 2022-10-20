package build

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	apiV2Group := engine.Group("/apis/front/v2")
	apiV2Route := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "buildschema",
			HandlerFunc: api.Get,
		},
	}
	route.RegisterRoutes(apiV2Group, apiV2Route)
}
