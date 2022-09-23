package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/cluster"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

type API struct {
	clusterCtl cluster.Controller
}

func NewAPI(clusterCtl cluster.Controller) *API {
	return &API{
		clusterCtl: clusterCtl,
	}
}

func (a *API) Create(c *gin.Context) {
	op := "cluster: create v2"
	applicationIDStr := c.Param(common.ParamApplicationID)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	scope := c.Request.URL.Query().Get(common.ClusterQueryScope)
	log.Infof(c, "scope: %v", scope)
	scopeArray := strings.Split(scope, "/")
	if len(scopeArray) != 2 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid scope!")
		return
	}
	environment := scopeArray[0]
	region := scopeArray[1]

	mergePatch := false
	mergePatchStr := c.Request.URL.Query().Get(common.ClusterQueryMergePatch)
	if mergePatchStr != "" {
		mergePatch, err = strconv.ParseBool(mergePatchStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("mergepatch is invalid, err: %v", err))
			return
		}
	}

	extraOwners := c.QueryArray(common.ClusterQueryExtraOwner)

	var request *cluster.CreateClusterRequestV2
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

	resp, err := a.clusterCtl.CreateClusterV2(c, uint(applicationID), environment,
		region, request, mergePatch)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ApplicationInDB {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrNameConflict {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, request)
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
	op := "cluster: update v2"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	mergePatch := false
	mergepatchStr := c.Request.URL.Query().Get(common.ClusterQueryMergePatch)
	if mergepatchStr != "" {
		mergePatch, err = strconv.ParseBool(mergepatchStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("mergepatch is invalid, err: %v", err))
			return
		}
	}

	var request *cluster.UpdateClusterRequestV2
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	err = a.clusterCtl.UpdateClusterV2(c, uint(clusterID), request, mergePatch)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}

		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Warningf("err = %+v, request = %+v", err, request)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}

		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

// Get V2 get api can also be used to get original cluster
func (a *API) Get(c *gin.Context) {
	op := "cluster: get v2"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	resp, err := a.clusterCtl.GetClusterV2(c, uint(clusterID))
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

func (a *API) InternalDeploy(c *gin.Context) {
	op := "cluster: internal deploy v2"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	pipelinerunIDStr := c.Param(common.ParamPipelinerunID)
	pipelinerunID, err := strconv.ParseUint(pipelinerunIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.InternalDeployV2(c, uint(clusterID), uint(pipelinerunID), request)
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
