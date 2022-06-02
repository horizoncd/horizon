package harbor

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/harbor"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_harborIDParam = "harborID"
)

type API struct {
	harborCtl harbor.Controller
}

func NewAPI() *API {
	return &API{harborCtl: harbor.Ctl}
}

func (a *API) ListAll(c *gin.Context) {
	harbors, err := a.harborCtl.ListAll(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, harbors)
}

func (a *API) Create(c *gin.Context) {
	var request *harbor.CreateHarborRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	id, err := a.harborCtl.Create(c, request)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) Update(c *gin.Context) {
	idStr := c.Param(_harborIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	var request *harbor.UpdateHarborRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	err = a.harborCtl.UpdateByID(c, uint(id), request)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) DeleteByID(c *gin.Context) {
	idStr := c.Param(_harborIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	err = a.harborCtl.DeleteByID(c, uint(id))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrHarborUsedByRegions {
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}
