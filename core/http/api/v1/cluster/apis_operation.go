package cluster

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/cluster"
	herrors "github.com/horizoncd/horizon/core/errors"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var rolloutGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "rollouts",
}

func (a *API) BuildDeploy(c *gin.Context) {
	op := "cluster: build deploy"
	var request *cluster.BuildDeployRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.BuildDeploy(c, uint(clusterID), request)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			} else if e.Source == herrors.GitlabResource {
				log.WithFiled(c, "op", op).Errorf("%+v", err)
				response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetDiff(c *gin.Context) {
	op := "cluster: get diff"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	targetTag := c.Query(common.ClusterQueryTargetTag)
	targetBranch := c.Query(common.ClusterQueryTargetBranch)
	targetCommit := c.Query(common.ClusterQueryTargetCommit)
	var refType, ref string

	if targetTag != "" {
		refType = codemodels.GitRefTypeTag
		ref = targetTag
	} else if targetBranch != "" {
		refType = codemodels.GitRefTypeBranch
		ref = targetBranch
	} else if targetCommit != "" {
		refType = codemodels.GitRefTypeCommit
		ref = targetCommit
	}

	resp, err := a.clusterCtl.GetDiff(c, uint(clusterID), refType, ref)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB || e.Source == herrors.GitlabResource {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

// Deprecated
func (a *API) ClusterStatus(c *gin.Context) {
	op := "cluster: cluster status"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	// nolint
	resp, err := a.clusterCtl.GetClusterStatus(c, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) PodEvents(c *gin.Context) {
	op := "cluster: pod events"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	podName := c.Query(common.ClusterQueryPodName)
	if podName == "" {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "podName is empty")
		return
	}

	resp, err := a.clusterCtl.GetPodEvents(c, uint(clusterID), podName)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			} else if e.Source == herrors.PodsInK8S {
				log.WithFiled(c, "op", op).Errorf("%+v", err)
				response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) InternalDeploy(c *gin.Context) {
	op := "cluster: internal deploy"
	clusterIDStr := c.Param(common.ParamClusterID)
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
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			} else if e.Source == herrors.PipelinerunInDB {
				log.WithFiled(c, "op", op).Errorf("%+v", err)
				response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

func (a *API) Restart(c *gin.Context) {
	op := "cluster: restart"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.Restart(c, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Deploy(c *gin.Context) {
	op := "cluster: deploy"
	clusterIDStr := c.Param(common.ParamClusterID)
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
		switch e := perror.Cause(err).(type) {
		case *herrors.HorizonErrNotFound:
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}

		if perror.Cause(err) == herrors.ErrClusterNoChange || perror.Cause(err) == herrors.ErrShouldBuildDeployFirst {
			log.WithFiled(c, "op", op).Errorf("%+v", err)
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}

		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

// Deprecated: use ExecuteAction instead
func (a *API) Next(c *gin.Context) {
	op := "cluster: op"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	err = a.clusterCtl.ExecuteAction(c, uint(clusterID), "promote", rolloutGVR)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		if e := perror.Cause(err); e == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

const defaultTailLines = 1000

func (a *API) GetContainerLog(c *gin.Context) {
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var tailLines int64 = defaultTailLines
	tailLinesStr := c.Query(common.ClusterQueryTailLines)
	if tailLinesStr != "" {
		tailLinesUint64, err := strconv.ParseUint(tailLinesStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return
		}
		tailLines = int64(tailLinesUint64)
	}

	podName := c.Query(common.ClusterQueryPodName)
	containerName := c.Query(common.ClusterQueryContainerName)

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

// Deprecated
func (a *API) Online(c *gin.Context) {
	op := "cluster: online"
	clusterIDStr := c.Param(common.ParamClusterID)
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
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

// Deprecated
func (a *API) Offline(c *gin.Context) {
	op := "cluster: offline"
	clusterIDStr := c.Param(common.ParamClusterID)
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
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Exec(c *gin.Context) {
	op := "cluster: exec"
	clusterIDStr := c.Param(common.ParamClusterID)
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

	resp, err := a.clusterCtl.Exec(c, uint(clusterID), request)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Rollback(c *gin.Context) {
	op := "cluster: rollback"
	clusterIDStr := c.Param(common.ParamClusterID)
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
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetGrafanaDashBoard(c *gin.Context) {
	op := "cluster: get dashboard"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetGrafanaDashBoard(c, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) DeleteClusterPods(c *gin.Context) {
	op := "cluster: get cluster pods"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	pods := c.QueryArray(common.ClusterQueryPodName)
	if len(pods) == 0 {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("pod name list should not be empty"))
		return
	}

	resp, err := a.clusterCtl.DeleteClusterPods(c, uint(clusterID), pods)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

// Deprecated: use ExecuteAction instead
func (a *API) Promote(c *gin.Context) {
	const op = "cluster: promote"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perror.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.ExecuteAction(c, uint(clusterID), "promote-full", rolloutGVR)
	if err != nil {
		err = perror.Wrap(err, "failed to promote cluster")
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if e := perror.Cause(err); e == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

// Deprecated: use ExecuteAction instead
func (a *API) Pause(c *gin.Context) {
	const op = "cluster: pause"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perror.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.ExecuteAction(c, uint(clusterID), "pause", rolloutGVR)
	if err != nil {
		err = perror.Wrap(err, "failed to pause cluster")
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if e := perror.Cause(err); e == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

// Deprecated: use ExecuteAction instead
func (a *API) Resume(c *gin.Context) {
	const op = "cluster: resume"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perror.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.ExecuteAction(c, uint(clusterID), "resume", rolloutGVR)
	if err != nil {
		err = perror.Wrap(err, "failed to resume cluster")
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if e := perror.Cause(err); e == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

// Deprecated: for internal usage
func (a *API) Upgrade(c *gin.Context) {
	const op = "cluster: upgrade"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perror.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	err = a.clusterCtl.Upgrade(c, uint(clusterID))
	if err != nil {
		err = perror.Wrap(err, "failed to upgrade cluster")
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}
