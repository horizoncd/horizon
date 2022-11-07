package scope

import (
	"github.com/gin-gonic/gin"

	"g.hz.netease.com/horizon/core/controller/scope"
	"g.hz.netease.com/horizon/pkg/server/response"
)

type API struct {
	scopeCtrl scope.Controller
}

func NewAPI(controller scope.Controller) *API {
	return &API{scopeCtrl: controller}
}

func (a *API) ListRole(c *gin.Context) {
	scopes := a.scopeCtrl.ListScopes(c)
	response.SuccessWithData(c, scopes)
}
