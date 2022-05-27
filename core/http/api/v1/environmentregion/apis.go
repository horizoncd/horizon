package environmentregion

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/environmentregion"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_environmentRegionIDParam = "environmentRegionID"
)

type API struct {
	environmentRegionCtl environmentregion.Controller
}

func NewAPI() *API {
	return &API{environmentRegionCtl: environmentregion.Ctl}
}

func (a *API) ListAll(c *gin.Context) {
	regions, err := a.environmentRegionCtl.ListAll(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg("failed to list all environmentRegions"))
		return
	}

	response.SuccessWithData(c, regions)
}

func (a *API) Create(c *gin.Context) {
	var request *environmentregion.CreateEnvironmentRegionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	id, err := a.environmentRegionCtl.CreateEnvironmentRegion(c, request)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) SetDefault(c *gin.Context) {
	idStr := c.Param(_environmentRegionIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	err = a.environmentRegionCtl.SetEnvironmentRegionToDefault(c, uint(id))
	if err != nil {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
	}

	response.Success(c)
}

func (a *API) DeleteByID(c *gin.Context) {
	idStr := c.Param(_environmentRegionIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	err = a.environmentRegionCtl.DeleteByID(c, uint(id))
	if err != nil {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
	}

	response.Success(c)
}
