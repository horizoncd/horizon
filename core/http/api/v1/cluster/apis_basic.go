package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/cluster"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_scope         = "scope"
	_mergePatch    = "mergePatch"
	_environment   = "environment"
	_targetBranch  = "targetBranch"
	_targetCommit  = "targetCommit"
	_targetTag     = "targetTag"
	_containerName = "containerName"
	_podName       = "podName"
	_tailLines     = "tailLines"
	_start         = "start"
	_end           = "end"
	_extraOwner    = "extraOwner"
	_hard          = "hard"
)

type API struct {
	clusterCtl cluster.Controller
}

func NewAPI(clusterCtl cluster.Controller) *API {
	return &API{
		clusterCtl: clusterCtl,
	}
}

func (a *API) List(c *gin.Context) {
	const op = "cluster: list"
	applicationIDStr := c.Param(common.ParamApplicationID)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	filter := c.Query(common.Filter)
	environments := c.QueryArray(_environment)

	var (
		pageNumber, pageSize int
	)
	pageNumberStr := c.Query(common.PageNumber)
	if pageNumberStr == "" {
		pageNumber = common.DefaultPageNumber
	} else {
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageNumber")
			return
		}
	}

	pageSizeStr := c.Query(common.PageSize)
	if pageSizeStr == "" {
		pageSize = common.DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageSize")
			return
		}
	}

	var tagSelectors []models.TagSelector
	tsFromCtx, ok := c.Get(common.TagSelector)
	if ok {
		tagSelectors, ok = tsFromCtx.([]models.TagSelector)
		if !ok {
			log.WithFiled(c, "op", op).Errorf("%+v",
				perror.WithMessagef(herrors.ErrParamInvalid, "invalid tag selector: %v", tagSelectors))
		}
	}

	count, respList, err := a.clusterCtl.ListCluster(c, uint(applicationID), environments, filter, &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, tagSelectors)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: respList,
	})
}

func (a *API) Create(c *gin.Context) {
	op := "cluster: create"
	applicationIDStr := c.Param(common.ParamApplicationID)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	scope := c.Request.URL.Query().Get(_scope)
	log.Infof(c, "scope: %v", scope)
	scopeArray := strings.Split(scope, "/")
	if len(scopeArray) != 2 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid scope!")
		return
	}
	environment := scopeArray[0]
	region := scopeArray[1]

	mergePatch := false
	mergepatchStr := c.Request.URL.Query().Get(_mergePatch)
	if mergepatchStr != "" {
		mergePatch, err = strconv.ParseBool(mergepatchStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("mergepatch is invalid, err: %v", err))
			return
		}
	}

	extraOwners := c.QueryArray(_extraOwner)

	var request *cluster.CreateClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	for _, roleOfMember := range request.ExtraMembers {
		if !role.CheckRoleIfValid(roleOfMember) {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("extra member is invalid"))
			return
		}
	}

	if request.ExtraMembers == nil {
		request.ExtraMembers = make(map[string]string)
	}
	for _, extraOwner := range extraOwners {
		request.ExtraMembers[extraOwner] = role.Owner
	}
	resp, err := a.clusterCtl.CreateCluster(c, uint(applicationID), environment,
		region, request, mergePatch)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ApplicationInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Errorf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Errorf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Update(c *gin.Context) {
	op := "cluster: update"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	mergePatch := false
	mergepatchStr := c.Request.URL.Query().Get(_mergePatch)
	if mergepatchStr != "" {
		mergePatch, err = strconv.ParseBool(mergepatchStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("mergepatch is invalid, err: %v", err))
			return
		}
	}

	var request *cluster.UpdateClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.UpdateCluster(c, uint(clusterID), request, mergePatch)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}

		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Errorf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}

		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Get(c *gin.Context) {
	op := "cluster: get"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	resp, err := a.clusterCtl.GetCluster(c, uint(clusterID))
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

func (a *API) Delete(c *gin.Context) {
	op := "cluster: delete"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	hard := false
	hardStr, ok := c.GetQuery(_hard)
	if ok {
		hard, err = strconv.ParseBool(hardStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return
		}
	}
	if err := a.clusterCtl.DeleteCluster(c, uint(clusterID), hard); err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}

		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) ListByNameFuzzily(c *gin.Context) {
	op := "cluster: list by name fuzzily"
	filter := c.Query(common.Filter)
	environment := c.Query(_environment)

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	keywords := request.GetFilterParam(c)

	count, respList, err := a.clusterCtl.ListClusterByNameFuzzily(c, environment, filter, &q.Query{
		Keywords:   keywords,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: respList,
	})
}

func (a *API) ListUserClusterByNameFuzzily(c *gin.Context) {
	op := "cluster: list user cluster by name fizzily"
	filter := c.Query(common.Filter)
	environment := c.Query(_environment)

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	keywords := request.GetFilterParam(c)

	count, respList, err := a.clusterCtl.ListUserClusterByNameFuzzily(c, environment, filter, &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Keywords:   keywords,
	})
	if err != nil {
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: respList,
	})
}

func (a *API) Free(c *gin.Context) {
	op := "cluster: free"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	err = a.clusterCtl.FreeCluster(c, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) GetOutput(c *gin.Context) {
	op := "cluster: get output"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	outPut, err := a.clusterCtl.GetClusterOutput(c, uint(clusterID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, outPut)
}

func (a *API) GetContainers(c *gin.Context) {
	const op = "cluster: get containers"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	podName := c.Query(_podName)
	if podName == "" {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("podName should not be empty"))
		return
	}

	outPut, err := a.clusterCtl.GetContainers(c, uint(clusterID), podName)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ClusterInDB || e.Source == herrors.ApplicationInDB || e.Source == herrors.PodsInK8S {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, outPut)
}

func (a *API) GetClusterPod(c *gin.Context) {
	op := "cluster: get cluster pod"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	podName := c.Query(_podName)
	if podName == "" {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("podName should not be empty"))
		return
	}

	resp, err := a.clusterCtl.GetClusterPod(c, uint(clusterID), podName)
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

func (a *API) GetByName(c *gin.Context) {
	op := "cluster: get by name"
	clusterName := c.Param(common.ParamClusterName)
	resp, err := a.clusterCtl.GetClusterByName(c, clusterName)
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
