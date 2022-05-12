package role

import (
	"g.hz.netease.com/horizon/core/controller/role"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

type API struct {
	roleCtrl role.Controller
}

func NewAPI(controller role.Controller) *API {
	return &API{roleCtrl: controller}
}

func (a *API) ListRole(c *gin.Context) {
	roles, err := a.roleCtrl.ListRegions(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, roles)
}
