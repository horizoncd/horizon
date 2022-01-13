package access

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	frontGroup := engine.Group("/apis/front/v1")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/accessreview",
			HandlerFunc: api.AccessReview,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
}
