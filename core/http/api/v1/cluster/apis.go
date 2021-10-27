package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/cluster"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_applicationIDParam = "applicationID"
	_scope              = "scope"
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
	applicationIDStr := c.Param(_applicationIDParam)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	scope := c.Query(_scope)
	scopeArray := strings.Split(scope, "/")
	if len(scopeArray) != 2 {
		response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid scope!")
		return
	}
	environment := scopeArray[0]
	region := scopeArray[1]

	var request *cluster.CreateClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.clusterCtl.CreateCluster(c, uint(applicationID), environment, region, request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}
