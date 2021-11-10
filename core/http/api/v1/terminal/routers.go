package terminal

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/front/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/terminal/sessionid", _clusterIDParam),
			HandlerFunc: api.GetSessionID,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/terminal/sockjs/*name",
			HandlerFunc: api.BindSockJs,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
