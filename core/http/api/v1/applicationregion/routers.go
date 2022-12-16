package applicationregion

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	apiGroup := engine.Group("/apis/core/v1")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/applications/:%v/defaultregions", common.ParamApplicationID),
			HandlerFunc: api.List,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/applications/:%v/defaultregions", common.ParamApplicationID),
			HandlerFunc: api.Update,
		},
	}

	route.RegisterRoutes(apiGroup, routes)
}
