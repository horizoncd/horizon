package cloudevent

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/server/route"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/internal")
	var routes = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/cloudevents",
			HandlerFunc: api.CloudEvent,
		},
	}
	route.RegisterRoutes(group, routes)
}
