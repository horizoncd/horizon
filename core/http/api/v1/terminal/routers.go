package terminal

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes.
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v1")
	coreRoutes := route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/terminal", _clusterIDParam),
			HandlerFunc: api.CreateTerminal,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/shell", _clusterIDParam),
			HandlerFunc: api.CreateShell,
		},
	}

	frontGroup := engine.Group("/apis/front/v1")
	frontRoutes := route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/terminal/:%v/websocket", _terminalIDParam),
			HandlerFunc: api.ConnectTerminal,
		},
	}

	route.RegisterRoutes(coreGroup, coreRoutes)
	route.RegisterRoutes(frontGroup, frontRoutes)
}
