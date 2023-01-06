package scope

import (
	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/controller/scope"
	"github.com/horizoncd/horizon/pkg/server/response"
)

type API struct {
	scopeCtrl scope.Controller
}

func NewAPI(controller scope.Controller) *API {
	return &API{scopeCtrl: controller}
}

func (a *API) ListScopes(c *gin.Context) {
	scopes := a.scopeCtrl.ListScopes(c)
	response.SuccessWithData(c, scopes)
}
