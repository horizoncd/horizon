package terminal

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoutes(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v2")
	var coreRoutes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/shell", _clusterIDParam),
			HandlerFunc: api.CreateShell,
		},
	}
	route.RegisterRoutes(coreGroup, coreRoutes)
}
