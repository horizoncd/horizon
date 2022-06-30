package tag

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/tag"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	_resourceTypeParam = "resourceType"
	_resourceIDParam   = "resourceID"
)

type API struct {
	tagCtl tag.Controller
}

func NewAPI(tagCtl tag.Controller) *API {
	return &API{
		tagCtl: tagCtl,
	}
}

func (a *API) List(c *gin.Context) {
	const op = "tag: list"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	resp, err := a.tagCtl.List(c, resourceType, uint(resourceID))
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

func (a *API) Update(c *gin.Context) {
	const op = "tag: update"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	var request *tag.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid request body, err: %s", err.Error())))
		return
	}
	err = a.tagCtl.Update(c, resourceType, uint(resourceID), request)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) ListSubResourceTags(c *gin.Context) {
	const op = "tag: list sub resource tags"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	resp, err := a.tagCtl.ListSubResourceTags(c, resourceType, uint(resourceID))
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
