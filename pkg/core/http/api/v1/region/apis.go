package region

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/region"
	"github.com/horizoncd/horizon/pkg/core/controller/tag"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	// param
	_regionIDParam = "id"
)

type API struct {
	regionCtl region.Controller
	tagCtl    tag.Controller
}

func NewAPI(ctl region.Controller, tagCtl tag.Controller) *API {
	return &API{
		regionCtl: ctl,
		tagCtl:    tagCtl,
	}
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
		if perror.Cause(err) == herrors.ErrRegistryUsedByRegions {
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) GetByID(c *gin.Context) {
	regionIDStr := c.Param(_regionIDParam)
	regionID, err := strconv.ParseUint(regionIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid regionID: %s, err: %s",
			regionIDStr, err.Error())))
		return
	}

	regionEntity, err := a.regionCtl.GetByID(c, uint(regionID))
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

func (a *API) ListRegionTags(c *gin.Context) {
	const op = "tag: list"
	regionIDStr := c.Param(_regionIDParam)
	regionID, err := strconv.ParseUint(regionIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", regionIDStr)))
		return
	}

	resp, err := a.tagCtl.List(c, common.ResourceRegion, uint(regionID))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}
