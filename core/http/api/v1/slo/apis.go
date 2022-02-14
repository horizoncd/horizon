package slo

import (
	"g.hz.netease.com/horizon/core/controller/slo"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

const (
	_internal = "internal"
	_env      = "env"
)

type API struct {
	sloController slo.Controller
}

func NewAPI(sloController slo.Controller) *API {
	return &API{sloController: sloController}
}

func (a *API) getAPIDashboard(c *gin.Context) {
	internal := c.Query(_internal)
	dashboard, err := a.sloController.GetAPIDashboard(c, internal)
	if err != nil {
		if perrors.Cause(err) == slo.ErrInternalInValid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, dashboard)
}

func (a *API) getPipelineDashboard(c *gin.Context) {
	internal := c.Query(_internal)
	env := c.Query(_env)
	dashboard, err := a.sloController.GetPipelineDashboard(c, internal, env)
	if err != nil {
		if perrors.Cause(err) == slo.ErrInternalInValid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, dashboard)
}

func (a *API) getOverviewDashboard(c *gin.Context) {
	env := c.Query(_env)
	dashboard := a.sloController.GetOverviewDashboard(c, env)

	response.SuccessWithData(c, dashboard)
}
