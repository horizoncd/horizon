package tag

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

func (api *API) RegisterRoute(engine *gin.Engine) {
	group := engine.Group("/apis/core/v1")
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
