package slo

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	frontGroup := engine.Group("/apis/front/v1")
	var frontSlos = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/slo/apidashboards",
			HandlerFunc: api.getDashboards,
		},
	}

	route.RegisterRoutes(frontGroup, frontSlos)
}
