package templateschematag

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/templateschematag"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	_clusterIDParam = "clusterID"
)

type API struct {
	templateSchemaTagCtl templateschematag.Controller
}

func NewAPI(tagCtl templateschematag.Controller) *API {
	return &API{
		templateSchemaTagCtl: tagCtl,
	}
}

func (a *API) List(c *gin.Context) {
	const op = "template schema tag: list"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid cluster id: %s", clusterIDStr)))
		return
	}

	resp, err := a.templateSchemaTagCtl.List(c, uint(clusterID))
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
	const op = "template schema tag: update"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid cluster id: %s", clusterIDStr)))
		return
	}

	var request *templateschematag.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid body: %s, err: %s", clusterIDStr, err.Error())))
		return
	}
	err = a.templateSchemaTagCtl.Update(c, uint(clusterID), request)
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
