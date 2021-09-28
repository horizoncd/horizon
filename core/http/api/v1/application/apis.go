package application

import (
	"fmt"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/controller/application"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_groupIDParam     = "groupID"
	_applicationParam = "application"
)

type API struct {
	applicationCtl application.Controller
}

func NewAPI(applicationCtl application.Controller) *API {
	return &API{
		applicationCtl: applicationCtl,
	}
}

func (a *API) Get(c *gin.Context) {
	name := c.Param(_applicationParam)
	var res *application.GetApplicationResponse
	var err error
	if res, err = a.applicationCtl.GetApplication(c, name); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, res)
}

func (a *API) Create(c *gin.Context) {
	var request *application.CreateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	if err := a.applicationCtl.CreateApplication(c, request); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

func (a *API) Update(c *gin.Context) {
	var request *application.UpdateApplicationRequest
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
