package pipelinestats

import (
	"g.hz.netease.com/horizon/core/common"
	pipelinestats "g.hz.netease.com/horizon/core/controller/pipelinestats"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	_application = "application"
	_cluster     = "cluster"
)

type API struct {
	pipelineStats pipelinestats.Controller
}

func NewAPI(c pipelinestats.Controller) *API {
	return &API{
		pipelineStats: c,
	}
}

func (a *API) GetApplicationPipelineStats(c *gin.Context) {
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	application := c.Query(_application)
	if application == "" {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("required application field is empty"))
		return
	}

	cluser := c.Query(_cluster)

	pipelineStats, count, err := a.pipelineStats.GetApplicationPipelineStats(c, application, cluser, pageNumber, pageSize)
	if err != nil {
		log.Errorf(c, "Get application pipelineStats failed, error: %+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: pipelineStats,
	})
}
