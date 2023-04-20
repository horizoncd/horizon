package cluster

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/cluster"
	herrors "github.com/horizoncd/horizon/core/errors"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/log"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const JWTTokenHeader = "X-Horizon-JWT-Token"

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

func (a *API) GetResourceTree(c *gin.Context) {
	op := "cluster: get resource tree"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetResourceTree(c, uint(clusterID))
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

func (a *API) GetStep(c *gin.Context) {
	op := "cluster: get step"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetStep(c, uint(clusterID))
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

func (a *API) ClusterStatus(c *gin.Context) {
	op := "cluster: cluster status"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetClusterStatusV2(c, uint(clusterID))
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

func (a *API) ClusterBuildStatus(c *gin.Context) {
	op := "cluster: cluster build status"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.clusterCtl.GetClusterBuildStatus(c, uint(clusterID))
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
	op := "cluster: internal deploy v2"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *cluster.InternalDeployRequestV2
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	tokenString := c.Request.Header.Get(JWTTokenHeader)
	if tokenString == "" {
		response.AbortWithUnauthorized(c, common.Unauthorized,
			"jwt token is empty!")
		return
	}
	var ctx context.Context = c
	ctx = common.WithContextJWTTokenString(ctx, tokenString)

	resp, err := a.clusterCtl.InternalDeployV2(ctx, uint(clusterID), request)
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
		if perror.Cause(err) == herrors.ErrTokenInvalid {
			log.WithFiled(c, "op", op).Errorf("%+v", err)
			response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			return
		}
		if perror.Cause(err) == herrors.ErrForbidden {
			log.WithFiled(c, "op", op).Errorf("%+v", err)
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, resp)
}

func (a *API) InternalClusterStatus(c *gin.Context) {
	op := "cluster: internal cluster status"

	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	tokenString := c.Request.Header.Get(JWTTokenHeader)
	if tokenString == "" {
		response.AbortWithUnauthorized(c, common.Unauthorized,
			"jwt token is empty!")
	}
	var ctx context.Context = c
	ctx = common.WithContextJWTTokenString(ctx, tokenString)

	resp, err := a.clusterCtl.InternalGetClusterStatus(ctx, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrTokenInvalid {
			log.WithFiled(c, "op", op).Errorf("%+v", err)
			response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithError(c, err)
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

func (a *API) ExecuteAction(c *gin.Context) {
	const op = "cluster: execute action"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		err = perror.Wrap(err, "failed to parse cluster id")
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("invalid cluster id"))
		return
	}

	var request cluster.ExecuteActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	err = a.clusterCtl.ExecuteAction(c, uint(clusterID), request.Action, schema.GroupVersionResource{
		Group:    request.Group,
		Version:  request.Version,
		Resource: request.Resource,
	})
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
