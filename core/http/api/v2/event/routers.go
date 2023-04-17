package event

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes.
func (a *API) RegisterRoute(engine *gin.Engine) {
	coreAPI := engine.Group("/apis/core/v2")
	coreRoutes := route.Routes{
		{
			Pattern:     "/supportevents",
			Method:      http.MethodGet,
			HandlerFunc: a.ListSupportEvents,
		},
	}

	route.RegisterRoutes(coreAPI, coreRoutes)
}
