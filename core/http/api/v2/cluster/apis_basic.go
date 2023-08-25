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

package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/cluster"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
	tagutil "github.com/horizoncd/horizon/pkg/util/tag"
)

type API struct {
	clusterCtl cluster.Controller
}

func NewAPI(clusterCtl cluster.Controller) *API {
	return &API{
		clusterCtl: clusterCtl,
	}
}

func parseContext(c *gin.Context) *q.Query {
	keywords := make(map[string]interface{})

	userIDStr := c.Query(common.ClusterQueryByUser)
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsgf(
					"failed to parse userID\n"+
						"userID = %s\nerr = %v", userIDStr, err))
			return nil
		}
		keywords[common.ClusterQueryByUser] = uint(userID)
	}

	applicationIDStr := c.Param(common.ParamApplicationID)
	if applicationIDStr != "" {
		applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
			return nil
		}
		keywords[common.ParamApplicationID] = applicationID
	}

	region := c.Query(common.ClusterQueryByRegion)
	if region != "" {
		keywords[common.ClusterQueryByRegion] = region
	}

	withFavorite := c.Query(common.ClusterQueryWithFavorite)
	if withFavorite != "" {
		keywords[common.ClusterQueryWithFavorite] = withFavorite
	}

	isFavoriteStr := c.Query(common.ClusterQueryIsFavorite)
	if isFavoriteStr != "" {
		isFavorite, err := strconv.ParseBool(isFavoriteStr)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return nil
		}
		keywords[common.ClusterQueryIsFavorite] = isFavorite
	}

	filter := c.Query(common.ClusterQueryName)
	if filter != "" {
		keywords[common.ClusterQueryName] = filter
	}

	environments := c.QueryArray(common.ClusterQueryEnvironment)
	if len(environments) == 1 && environments[0] != "" {
		keywords[common.ClusterQueryEnvironment] = environments[0]
	} else if len(environments) > 1 {
		keywords[common.ClusterQueryEnvironment] = environments
	}

	tagSelectorStr := c.Query(common.ClusterQueryTagSelector)
	if tagSelectorStr != "" {
		tagSelectors, err := tagutil.ParseTagSelector(tagSelectorStr)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsgf(
					"failed to parse tagSelector\n"+
						"selector = %s\nerr = %v", tagSelectorStr, err))
			return nil
		}
		keywords[common.ClusterQueryTagSelector] = tagSelectors
	}

	template := c.Query(common.ClusterQueryByTemplate)
	if template != "" {
		keywords[common.ClusterQueryByTemplate] = template
	}

	release := c.Query(common.ClusterQueryByRelease)
	if release != "" {
		keywords[common.ClusterQueryByRelease] = release
	}

	query := q.New(keywords).WithPagination(c)
	return query
}

func (a *API) ListSelf(c *gin.Context) {
	currentUser, err := common.UserFromContext(c)
	if err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf(
				"current user not found\n"+
					"err = %v", err))
		return
	}

	c.Request.URL.RawQuery =
		fmt.Sprintf("%s%s", c.Request.URL.RawQuery,
			fmt.Sprintf("&%s=%d", common.ClusterQueryByUser, currentUser.GetID()))
	a.List(c)
}

func (a *API) List(c *gin.Context) {
	const op = "cluster: list"

	query := parseContext(c)
	if query == nil {
		return
	}

	respList, count, err := a.clusterCtl.List(c, query)
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

func (a *API) ListByApplication(c *gin.Context) {
	const op = "cluster: list"
	query := parseContext(c)

	count, respList, err := a.clusterCtl.ListByApplication(c, query)
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

	resp, err := a.clusterCtl.CreateClusterV2(c, &cluster.CreateClusterParamsV2{
		CreateClusterRequestV2: request,
		ApplicationID:          uint(applicationID),
		Environment:            environment,
		Region:                 region,
		MergePatch:             mergePatch,
	})
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

func (a *API) Delete(c *gin.Context) {
	op := "cluster: delete"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	hard := false
	hardStr, ok := c.GetQuery(common.ClusterQueryHard)
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
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
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

	podName := c.Query(common.ClusterQueryPodName)
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

	podName := c.Query(common.ClusterQueryPodName)
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

func (a *API) AddFavorite(c *gin.Context) {
	op := "cluster: add favorite"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	whetherLike := cluster.WhetherLike{IsFavorite: true}

	err = a.clusterCtl.ToggleLikeStatus(c, uint(clusterID), &whetherLike)
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

func (a *API) DeleteFavorite(c *gin.Context) {
	op := "cluster: delete favorite"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	whetherLike := cluster.WhetherLike{IsFavorite: false}

	err = a.clusterCtl.ToggleLikeStatus(c, uint(clusterID), &whetherLike)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(e.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) CreatePipelineRun(c *gin.Context) {
	op := "cluster: create pipeline run"
	clusterIDStr := c.Param(common.ParamClusterID)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var req cluster.CreatePipelineRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	pipelineRun, err := a.clusterCtl.CreatePipelineRun(c, uint(clusterID), &req)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, pipelineRun)
}
