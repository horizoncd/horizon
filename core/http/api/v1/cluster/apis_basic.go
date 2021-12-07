package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/cluster"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_applicationIDParam = "applicationID"
	_clusterIDParam     = "clusterID"
	_clusterParam       = "cluster"
	_scope              = "scope"
	_environment        = "environment"
	_targetBranch       = "targetBranch"
	_containerName      = "containerName"
	_podName            = "podName"
	_tailLines          = "tailLines"
	_start              = "start"
	_end                = "end"
	_extraOwner         = "extraOwner"
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
	applicationIDStr := c.Param(_applicationIDParam)
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

	count, respList, err := a.clusterCtl.ListCluster(c, uint(applicationID), environments, filter, &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: respList,
	})
}

func (a *API) Create(c *gin.Context) {
	applicationIDStr := c.Param(_applicationIDParam)
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

	extraOwners := c.QueryArray(_extraOwner)

	// add query for migration
	// TODO(gjq): remove these two query params after migration
	namespace := c.Query("namespace")
	image := c.Query("image")

	var request *cluster.CreateClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.CreateCluster(c, uint(applicationID), environment, region, extraOwners, request,
		namespace, image)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Update(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *cluster.UpdateClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.UpdateCluster(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Get(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	resp, err := a.clusterCtl.GetCluster(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetByName(c *gin.Context) {
	clusterName := c.Param(_clusterParam)
	resp, err := a.clusterCtl.GetClusterByName(c, clusterName)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Delete(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	if err := a.clusterCtl.DeleteCluster(c, uint(clusterID)); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

func (a *API) ListByNameFuzzily(c *gin.Context) {
	filter := c.Query(common.Filter)
	environment := c.Query(_environment)

	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	count, respList, err := a.clusterCtl.ListClusterByNameFuzzily(c, environment, filter, &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(count),
		Items: respList,
	})
}

func (a *API) Free(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	err = a.clusterCtl.FreeCluster(c, uint(clusterID))
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}
