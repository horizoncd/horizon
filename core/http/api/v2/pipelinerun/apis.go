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

package pipelinerun

import (
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	prctl "github.com/horizoncd/horizon/core/controller/pipelinerun"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/request"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/errors"

	"github.com/gin-gonic/gin"
)

const (
	_pipelinerunIDParam = "pipelinerunID"
	_checkrunIDParam    = "checkrunID"
	_clusterIDParam     = "clusterID"
	_canRollbackParam   = "canRollback"
	_pipelineStatus     = "status"
)

type API struct {
	prCtl prctl.Controller
}

func NewAPI(prCtl prctl.Controller) *API {
	return &API{
		prCtl: prCtl,
	}
}

func (a *API) Log(c *gin.Context) {
	a.withPipelinerunID(c, func(prID uint) {
		l, err := a.prCtl.GetPipelinerunLog(c, uint(prID))
		if err != nil {
			l := &collector.Log{
				LogBytes: []byte(errors.Message(err)),
			}
			a.writeLog(c, l)
			return
		}
		a.writeLog(c, l)
	})
}

func (a *API) writeLog(c *gin.Context, l *collector.Log) {
	c.Header("Content-Type", "text/plain")
	if l.LogBytes != nil {
		_, _ = c.Writer.Write(l.LogBytes)
		return
	}

	logC := l.LogChannel
	errC := l.ErrChannel
	for logC != nil || errC != nil {
		select {
		case l, ok := <-logC:
			if !ok {
				logC = nil
				continue
			}
			if l.Log == "EOFLOG" {
				_, _ = c.Writer.Write([]byte("\n"))
				continue
			}
			_, _ = c.Writer.Write([]byte(fmt.Sprintf("[%s : %s] %s\n", l.Task, l.Step, l.Log)))
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			_, _ = c.Writer.Write([]byte(fmt.Sprintf("%s\n", e)))
		}
	}
}

func (a *API) GetDiff(c *gin.Context) {
	a.withPipelinerunID(c, func(pipelinerunID uint) {
		diff, err := a.prCtl.GetDiff(c, uint(pipelinerunID))
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, diff)
	})
}

func (a *API) Get(c *gin.Context) {
	a.withPipelinerunID(c, func(pipelinerunID uint) {
		resp, err := a.prCtl.GetPipelinerun(c, uint(pipelinerunID))
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, resp)
	})
}

func (a *API) List(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	canRollbackStr := c.Query(_canRollbackParam)
	canRollback, err := strconv.ParseBool(canRollbackStr)
	if err != nil {
		canRollback = false
	}
	status := c.QueryArray(_pipelineStatus)
	keywords := map[string]interface{}{}
	if len(status) > 0 {
		keywords[common.PipelineQueryByStatus] = status
	}

	query := q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Keywords:   keywords,
	}
	total, pipelines, err := a.prCtl.ListPipelineruns(c, uint(clusterID), canRollback, query)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: pipelines,
	})
}

func (a *API) Stop(c *gin.Context) {
	a.withPipelinerunID(c, func(prID uint) {
		err := a.prCtl.StopPipelinerun(c, uint(prID))
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.Success(c)
	})
}

func (a *API) ExecuteForce(c *gin.Context) {
	a.execute(c, true)
}

func (a *API) Execute(c *gin.Context) {
	a.execute(c, false)
}

func (a *API) execute(c *gin.Context, force bool) {
	a.withPipelinerunID(c, func(prID uint) {
		err := a.prCtl.Execute(c, prID, force)
		if err != nil {
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(e.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
		response.Success(c)
	})
}

func (a *API) Cancel(c *gin.Context) {
	a.withPipelinerunID(c, func(prID uint) {
		err := a.prCtl.Cancel(c, prID)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.Success(c)
	})
}

func (a *API) ListCheckRuns(c *gin.Context) {
	query := parseContext(c)
	checkRuns, err := a.prCtl.ListCheckRuns(c, query)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, checkRuns)

}

func (a *API) CreateCheckRun(c *gin.Context) {
	var req prctl.CreateOrUpdateCheckRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	a.withPipelinerunID(c, func(prID uint) {
		checkrun, err := a.prCtl.CreateCheckRun(c, prID, &req)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, checkrun)
	})
}

func (a *API) GetCheckRun(c *gin.Context) {
	a.withCheckrunID(c, func(id uint) {
		checkrun, err := a.prCtl.GetCheckRunByID(c, id)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, checkrun)
	})
}

func (a *API) UpdateCheckRun(c *gin.Context) {
	var req prctl.CreateOrUpdateCheckRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	a.withCheckrunID(c, func(id uint) {
		err := a.prCtl.UpdateCheckRunByID(c, id, &req)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.Success(c)
	})
}

func (a *API) CreatePrMessage(c *gin.Context) {
	var req prctl.CreatePrMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	a.withPipelinerunID(c, func(prID uint) {
		checkMessage, err := a.prCtl.CreatePRMessage(c, prID, &req)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, checkMessage)
	})
}

func (a *API) ListPrMessages(c *gin.Context) {
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	query := &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}
	a.withPipelinerunID(c, func(prID uint) {
		count, messages, err := a.prCtl.ListPRMessages(c, prID, query)
		if err != nil {
			response.AbortWithError(c, err)
			return
		}
		response.SuccessWithData(c, response.DataWithTotal{
			Total: int64(count),
			Items: messages,
		})
	})
}

func (a *API) withPipelinerunID(c *gin.Context, f func(pipelineRunID uint)) {
	idStr := c.Param(_pipelinerunIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	f(uint(id))
}

func (a *API) withCheckrunID(c *gin.Context, f func(pipelineRunID uint)) {
	idStr := c.Param(_checkrunIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	f(uint(id))
}

func parseContext(c *gin.Context) *q.Query {
	keywords := make(map[string]interface{})

	filter := c.Query(common.CheckrunQueryFilter)
	if filter != "" {
		keywords[common.CheckrunQueryFilter] = filter
	}

	status := c.Query(common.CheckrunQueryByStatus)
	if status != "" {
		keywords[common.CheckrunQueryByStatus] = status
	}

	pipelinerunIDStr := c.Query(common.CheckrunQueryByPipelinerunID)
	if pipelinerunIDStr != "" {
		pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return nil
		}
		keywords[common.CheckrunQueryByPipelinerunID] = pipelinerunID
	}

	checkIDStr := c.Query(common.CheckrunQueryByCheckID)
	if checkIDStr != "" {
		checkID, err := strconv.ParseUint(checkIDStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return nil
		}
		keywords[common.CheckrunQueryByCheckID] = checkID
	}
	return q.New(keywords)
}
