package environment

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/environment"
	"g.hz.netease.com/horizon/core/controller/environmentregion"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_environmentParam   = "environment"
	_environmentIDParam = "environmentID"
)

type API struct {
	envCtl       environment.Controller
	envRegionCtl environmentregion.Controller
}

func NewAPI() *API {
	return &API{
		envCtl:       environment.Ctl,
		envRegionCtl: environmentregion.Ctl,
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
	regions, err := a.envRegionCtl.ListRegionsByEnvironment(c, env)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, regions)
}

func (a *API) Update(c *gin.Context) {
	const op = "environment: update"

	envIDStr := c.Param(_environmentIDParam)
	envID, err := strconv.ParseUint(envIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	var request *environment.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	err = a.envCtl.UpdateByID(c, uint(envID), request)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}
