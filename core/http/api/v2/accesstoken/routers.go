package accesstoken

import (
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/route"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func (api *API) RegisterRoute(engine *gin.Engine) {
	coreGroup := engine.Group("/apis/core/v2")
	var coreRouters = route.Routes{
		{
			Method:      http.MethodPost,
			Pattern:     "/personalaccesstokens",
			HandlerFunc: api.CreatePersonalAccessToken,
		},
		{
			Method:      http.MethodPost,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.CreateResourceAccessToken,
		},
		{
			Method:      http.MethodGet,
			Pattern:     "/personalaccesstokens",
			HandlerFunc: api.ListPersonalAccessTokens,
		},
		{
			Method:      http.MethodGet,
			Pattern:     fmt.Sprintf("/:%s/:%s/accesstokens", common.ParamResourceType, common.ParamResourceID),
			HandlerFunc: api.ListResourceAccessTokens,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/personalaccesstokens/:%s", common.ParamAccessTokenID),
			HandlerFunc: api.RevokePersonalAccessToken,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     fmt.Sprintf("/accesstokens/:%s", common.ParamAccessTokenID),
			HandlerFunc: api.RevokeResourceAccessToken,
		},
	}

	route.RegisterRoutes(coreGroup, coreRouters)
}
