package pipelinerun

import (
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/core/common"
	prctl "github.com/horizoncd/horizon/pkg/core/controller/pipelinerun"
	"github.com/horizoncd/horizon/pkg/server/request"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/errors"

	"github.com/gin-gonic/gin"
)

const (
	_pipelinerunIDParam = "pipelinerunID"
	_clusterIDParam     = "clusterID"
	_clusterParam       = "cluster"
	_canRollbackParam   = "canRollback"
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
	prIDStr := c.Param(_pipelinerunIDParam)
	prID, err := strconv.ParseUint(prIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	l, err := a.prCtl.GetPipelinerunLog(c, uint(prID))
	if err != nil {
		l := &prctl.Log{
			LogBytes: []byte(errors.Message(err)),
		}
		a.writeLog(c, l)
		return
	}
	a.writeLog(c, l)
}

func (a *API) LatestLogForCluster(c *gin.Context) {
	clusterIDStr := c.Param(_clusterParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	l, err := a.prCtl.GetClusterLatestLog(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	a.writeLog(c, l)
}

func (a *API) writeLog(c *gin.Context, l *prctl.Log) {
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
	pipelinerunIDStr := c.Param(_pipelinerunIDParam)
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	diff, err := a.prCtl.GetDiff(c, uint(pipelinerunID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, diff)
}

func (a *API) Get(c *gin.Context) {
	pipelinerunIDStr := c.Param(_pipelinerunIDParam)
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	resp, err := a.prCtl.Get(c, uint(pipelinerunID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
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

	total, pipelines, err := a.prCtl.List(c, uint(clusterID), canRollback, q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
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
	pipelinerunIDStr := c.Param(_pipelinerunIDParam)
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	err = a.prCtl.StopPipelinerun(c, uint(pipelinerunID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

func (a *API) StopPipelinerunForCluster(c *gin.Context) {
	clusterIDStr := c.Param(_clusterParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	err = a.prCtl.StopPipelinerunForCluster(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}
