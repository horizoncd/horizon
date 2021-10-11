package application

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	application2 "g.hz.netease.com/horizon/core/controller/application"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_groupIDParam     = "groupID"
	_applicationParam = "application"
)

type API struct {
	applicationCtl application2.Controller
}

func NewAPI(applicationCtl application2.Controller) *API {
	return &API{
		applicationCtl: applicationCtl,
	}
}

func (a *API) Get(c *gin.Context) {
	name := c.Param(_applicationParam)
	var res *application2.GetApplicationResponse
	var err error
	if res, err = a.applicationCtl.GetApplication(c, name); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, res)
}

func (a *API) Create(c *gin.Context) {
	groupIDStr := c.Param(_groupIDParam)
	groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	var request *application2.CreateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	if err := a.applicationCtl.CreateApplication(c, uint(groupID), request); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

func (a *API) Update(c *gin.Context) {
	var request *application2.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	name := c.Param(_applicationParam)
	if err := a.applicationCtl.UpdateApplication(c, name, request); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}
