package pipelinerun

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	_pipelinerunIDParam = "pipelinerunID"
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
	l, err := a.prCtl.GetPrLog(c, uint(prID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	c.Header("Content-Type", "text/plain")
	if l.LogBytes != nil {
		log.Infof(c, "logBytes: %v", string(l.LogBytes))
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
			log.Infof(c, "[%s : %s] %s\n", l.Task, l.Step, l.Log)
			_, _ = c.Writer.Write([]byte(fmt.Sprintf("[%s : %s] %s\n", l.Task, l.Step, l.Log)))
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			_, _ = c.Writer.Write([]byte(fmt.Sprintf("%s\n", e)))
		}
	}
	c.Writer.Flush()
}
