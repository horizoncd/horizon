// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	var wpr *cloudevent.WrappedPipelineRun
	if err := c.ShouldBindJSON(&wpr); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	if err := a.cloudEventCtl.CloudEvent(c, wpr); err != nil {
		log.Errorf(c, "failed to handle cloud event, pipelinerun name: %s, err: %v",
			wpr.PipelineRun.Name, err)
		response.AbortWithError(c, err)
		return
	}

	response.Success(c)
}
