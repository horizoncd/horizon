package templateschematag

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/route"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	group := engine.Group("/apis/core/v2")
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
