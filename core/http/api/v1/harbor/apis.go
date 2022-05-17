package harbor

import (
	"fmt"

	"g.hz.netease.com/horizon/core/controller/harbor"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

type API struct {
	harborCtl harbor.Controller
}

func NewAPI() *API {
	return &API{harborCtl: harbor.Ctl}
}

func (a *API) listAll(c *gin.Context) {
	regions, err := a.harborCtl.ListAll(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, regions)
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
