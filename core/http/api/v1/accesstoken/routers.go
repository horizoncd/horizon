package accesstoken

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine, api *API) {
	coreGroup := engine.Group("/apis/core/v1")
	var coreRouters = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/accesstokens",
			HandlerFunc: api.CreateAccessToken,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.CreateAccessToken,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/accesstokens",
			HandlerFunc: api.ListAccessTokens,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.ListAccessTokens,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/accesstokens/:%s", common.ParamAccessTokenID),
			HandlerFunc: api.RevokeAccessToken,
		},
	}

	route.RegisterRoutes(coreGroup, coreRouters)
}
