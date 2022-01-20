package access

import (
	"fmt"

	"g.hz.netease.com/horizon/core/controller/access"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

type API struct {
	accessCtl access.Controller
}

func NewAPI(c access.Controller) *API {
	return &API{
		accessCtl: c,
	}
}

func (a *API) AccessReview(c *gin.Context) {
	const op = "access: access review"

	var request *access.ReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	if len(request.APIs) == 0 {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("apis should not be empty"))
		return
	}

	reviewResp, err := a.accessCtl.Review(c, request.APIs)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, reviewResp)
}
