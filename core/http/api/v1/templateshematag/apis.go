package templateshematag

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/templateschema"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	_clusterIDParam = "clusterID"
)

type API struct {
	templateSchemaTagCtl templateschema.Controller
}

func NewAPI(clusterTagCtl templateschema.Controller) *API {
	return &API{
		templateSchemaTagCtl: clusterTagCtl,
	}
}

func (a *API) List(c *gin.Context) {
	clusterIDStr := c.Param(_clusterIDParam)
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	resp, err := a.templateSchemaTagCtl.List(c, uint(clusterID))
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

	var request *templateschema.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	err = a.templateSchemaTagCtl.Update(c, uint(clusterID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}
