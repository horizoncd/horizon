package role

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/core/controller/role"
	"github.com/horizoncd/horizon/pkg/server/response"
)

type API struct {
	roleCtrl role.Controller
}

func NewAPI(controller role.Controller) *API {
	return &API{roleCtrl: controller}
}

func (a *API) ListRole(c *gin.Context) {
	roles, err := a.roleCtrl.ListRole(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, roles)
}
