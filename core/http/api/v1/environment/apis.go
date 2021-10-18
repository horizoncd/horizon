package environment

import (
	"g.hz.netease.com/horizon/core/controller/environment"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_environmentParam = "environment"
)

type API struct {
	envCtl environment.Controller
}

func NewAPI() *API {
	return &API{
		envCtl: environment.Ctl,
	}
}

func (a *API) ListEnvironments(c *gin.Context) {
	envs, err := a.envCtl.ListEnvironments(c)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, envs)
}

func (a *API) ListEnvironmentRegions(c *gin.Context) {
	env := c.Param(_environmentParam)
	regions, err := a.envCtl.ListRegionsByEnvironment(c, env)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, regions)
}
