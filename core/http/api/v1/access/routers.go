package access

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes.
func (a *API) RegisterRoute(engine *gin.Engine) {
	frontGroup := engine.Group("/apis/front/v1")
	frontRoutes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/accessreview",
			HandlerFunc: a.AccessReview,
		},
	}

	route.RegisterRoutes(frontGroup, frontRoutes)
}
