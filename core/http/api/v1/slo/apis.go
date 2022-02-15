package slo

import (
	"g.hz.netease.com/horizon/core/controller/slo"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	_env = "env"
)

type API struct {
	sloController slo.Controller
}

func NewAPI(sloController slo.Controller) *API {
	return &API{sloController: sloController}
}

func (a *API) getDashboards(c *gin.Context) {
	env := c.Query(_env)

	response.SuccessWithData(c, a.sloController.GetDashboards(c, env))
}
