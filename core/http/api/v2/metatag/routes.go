package metatag

import (
	"github.com/horizoncd/horizon/pkg/server/route"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoute register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v1")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/metatags",
			HandlerFunc: api.CreateMetatags,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/metatags",
			HandlerFunc: api.GetMetatagsByKey,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/metatagkeys",
			HandlerFunc: api.GetMetatagKeys,
		},
	}

	route.RegisterRoutes(coreGroup, coreRoutes)
}
