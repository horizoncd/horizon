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

func (a *API) getOverviewDashboard(c *gin.Context) {
	env := c.Query(_env)
	dashboard := a.sloController.GetOverviewDashboard(c, env)

	response.SuccessWithData(c, dashboard)
}

func (a *API) getAPIDashboard(c *gin.Context) {
	dashboard := a.sloController.GetAPIDashboard(c)

	response.SuccessWithData(c, dashboard)
}

func (a *API) getPipelineDashboard(c *gin.Context) {
	env := c.Query(_env)
	dashboard := a.sloController.GetPipelineDashboard(c, env)

	response.SuccessWithData(c, dashboard)
}
