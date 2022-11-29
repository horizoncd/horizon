package event

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, a *API) {
	frontAPI := engine.Group("/apis/front/v1/eventactions")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			HandlerFunc: a.ListEventActions,
		},
	}

	route.RegisterRoutes(frontAPI, frontRoutes)
}
