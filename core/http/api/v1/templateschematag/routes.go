package templateschematag

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/clusters/:%v/templateschematags", _clusterIDParam),
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/clusters/:%v/templateschematags", _clusterIDParam),
			HandlerFunc: api.Update,
		},
	}
	route.RegisterRoutes(group, routes)
}
