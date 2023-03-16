package build

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (api *API) RegisterRoutes(engine *gin.Engine) {
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
