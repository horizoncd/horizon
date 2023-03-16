package access

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	frontGroup := engine.Group("/apis/front/v2")
	var frontRoutes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/accessreview",
			HandlerFunc: api.AccessReview,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
}
