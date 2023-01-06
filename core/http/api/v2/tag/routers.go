package tag

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, api *API) {
	group := engine.Group("/apis/core/v2")
	var routes = route.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/tags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.List,
		}, {
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%v/:%v/tags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.Update,
		}, {
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%v/:%v/subresourcetags", _resourceTypeParam, _resourceIDParam),
			HandlerFunc: api.ListSubResourceTags,
		},
	}
	route.RegisterRoutes(group, routes)
}
