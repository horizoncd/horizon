package cluster

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/cluster"
	"g.hz.netease.com/horizon/pkg/cluster/dao"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

func (a *API) BuildDeploy(c *gin.Context) {
	var request *cluster.BuildDeployRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.BuildDeploy(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetDiff(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	targetBranch := c.Query(_targetBranch)
	resp, err := a.clusterCtl.GetDiff(c, uint(clusterID), targetBranch)

	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) ClusterStatus(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetClusterStatus(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) PodEvents(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	podName := c.Query(_podName)
	if podName == "" {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "podName is empty")
		return
	}

	resp, err := a.clusterCtl.GetPodEvents(c, uint(clusterID), podName)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) InternalDeploy(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *cluster.InternalDeployRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.InternalDeploy(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Restart(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.Restart(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Deploy(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	var request *cluster.DeployRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	resp, err := a.clusterCtl.Deploy(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Next(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	err = a.clusterCtl.Next(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

const defaultTailLines = 1000

func (a *API) GetContainerLog(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	tailLines := defaultTailLines
	tailLinesStr := c.Query(_tailLines)
	if tailLinesStr != "" {
		tailLinesUint64, err := strconv.ParseUint(tailLinesStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return
		}
		tailLines = int(tailLinesUint64)
	}

	podName := c.Query(_podName)
	containerName := c.Query(_containerName)

	logC, err := a.clusterCtl.GetContainerLog(c, uint(clusterID), podName, containerName, tailLines)
	if err != nil {
		_, _ = c.Writer.Write([]byte(err.Error()))
		return
	}

	for logC != nil {
		l, ok := <-logC
		if !ok {
			logC = nil
			continue
		}
		_, _ = c.Writer.Write([]byte(l))
	}
}

func (a *API) Online(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *cluster.ExecRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	resp, err := a.clusterCtl.Online(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Offline(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *cluster.ExecRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	resp, err := a.clusterCtl.Offline(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Rollback(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	var request *cluster.RollbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	resp, err := a.clusterCtl.Rollback(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetDashBoard(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetDashboard(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetClusterPods(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	// start means start timestamp(seconds) of the query range
	startStr := c.Query(_start)
	start, err := strconv.ParseInt(startStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	// end means end timestamp(seconds) of the query range
	endStr := c.Query(_end)
	end, err := strconv.ParseInt(endStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetClusterPods(c, uint(clusterID), start, end)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Promote(c *gin.Context) {
	const op = "cluster: promote"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perrors.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.Promote(c, uint(clusterID))
	if err != nil {
		err = perrors.Wrap(err, "failed to pause cluster")
		if perrors.Cause(err) == dao.ErrClusterNotFound {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg("cluster not found"))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) Pause(c *gin.Context) {
	const op = "cluster: pause"
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perrors.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.Pause(c, uint(clusterID))
	if err != nil {
		err = perrors.Wrap(err, "failed to pause cluster")
		if perrors.Cause(err) == dao.ErrClusterNotFound {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg("cluster not found"))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}
