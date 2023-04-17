package cloudevent

import (
	"fmt"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/cloudevent"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
	pipelinecloudevent "github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"
)

type API struct {
	cloudEventCtl cloudevent.Controller
}

func NewAPI(cloudEventCtl cloudevent.Controller) *API {
	return &API{
		cloudEventCtl: cloudEventCtl,
	}
}

func (a *API) CloudEvent(c *gin.Context) {
	ceType := c.GetHeader("Ce-Type")
	if ceType != string(pipelinecloudevent.PipelineRunSuccessfulEventV1) &&
		ceType != string(pipelinecloudevent.PipelineRunFailedEventV1) {
		response.Success(c)
		return
	}

	wpr := cloudevent.WrappedPipelineRun{}
	if err := c.ShouldBindJSON(&wpr); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	if ceType == string(pipelinecloudevent.PipelineRunSuccessfulEventV1) {
		log.Infof(c, "pipeline run succeeded - pipelineRunName: %s", wpr.Name())
	} else if ceType == string(pipelinecloudevent.PipelineRunFailedEventV1) {
		log.Error(c, "pipeline run failed", "pipelineRunName", wpr.Name(), "reason", wpr.Reason(), "message", wpr.Message())
	}

	if err := a.cloudEventCtl.CloudEvent(c, &wpr); err != nil {
		log.Error(c, "failed to handle cloud event", "pipelineRunName", wpr.PipelineRun.Name, "err", err)
		response.AbortWithError(c, err)
		return
	}

	response.Success(c)
}
