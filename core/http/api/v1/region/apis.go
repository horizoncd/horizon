package region

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/region"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_regionIDParam   = "regionID"
	_regionNameParam = "regionName"
)

type API struct {
	regionCtl region.Controller
}

func NewAPI() *API {
	return &API{regionCtl: region.Ctl}
}

func (a *API) ListRegions(c *gin.Context) {
	regions, err := a.regionCtl.ListRegions(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, regions)
}

func (a *API) UpdateByID(c *gin.Context) {
	regionIDStr := c.Param(_regionIDParam)
	regionID, err := strconv.ParseUint(regionIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid regionID: %s, err: %s",
			regionIDStr, err.Error())))
		return
	}

	var request *region.UpdateRegionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	err = a.regionCtl.UpdateByID(c, uint(regionID), request)
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

func (a *API) Create(c *gin.Context) {
	var request *region.CreateRegionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	id, err := a.regionCtl.Create(c, request)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) DeleteByID(c *gin.Context) {
	regionIDStr := c.Param(_regionIDParam)
	regionID, err := strconv.ParseUint(regionIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid regionID: %s, err: %s",
			regionIDStr, err.Error())))
		return
	}

	err = a.regionCtl.DeleteByID(c, uint(regionID))
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

func (a *API) GetByName(c *gin.Context) {
	name := c.Param(_regionNameParam)
	if len(name) == 0 {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("region name cannot be empty"))
	}
	regionEntity, err := a.regionCtl.GetByName(c, name)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, regionEntity)
}
