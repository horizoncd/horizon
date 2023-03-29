package cloudevent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
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
