package application

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/application"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"

	"github.com/gin-gonic/gin"
)

const (
	// param
	_groupIDParam       = "groupID"
	_applicationIDParam = "applicationID"
	_extraOwner         = "extraOwner"
	_groupIDStr         = "groupID"
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
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	var res *application.GetApplicationResponse
	if res, err = a.applicationCtl.GetApplication(c, uint(appID)); err != nil {
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

	extraOwners := c.QueryArray(_extraOwner)

	var request *application.CreateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	resp, err := a.applicationCtl.CreateApplication(c, uint(groupID), extraOwners, request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Update(c *gin.Context) {
	var request *application.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	resp, err := a.applicationCtl.UpdateApplication(c, uint(appID), request)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Transfer(c *gin.Context) {
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	groupIDStr := c.Query(_groupIDStr)
	groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	err = a.applicationCtl.Transfer(c, uint(appID), uint(groupID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.GroupInDB {
				response.AbortWithRequestError(c, "GroupNotExist", err.Error())
				return
			} else if e.Source == herrors.ApplicationInDB {
				response.AbortWithNotExistError(c, err.Error())
				return
			}
		} else {
			response.AbortWithInternalError(c, err.Error())
			return
		}
	}
	response.Success(c)
}

func (a *API) Delete(c *gin.Context) {
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}
	if err := a.applicationCtl.DeleteApplication(c, uint(appID)); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}

// SearchApplication search all applications
func (a *API) SearchApplication(c *gin.Context) {
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	filter := c.Query(common.Filter)

	total, applications, err := a.applicationCtl.ListApplication(c, filter, q.Query{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	})
	if err != nil {
		response.AbortWithError(c, err)
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: applications,
	})
}

// SearchMyApplication search all applications that authorized to current user
func (a *API) SearchMyApplication(c *gin.Context) {
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	filter := c.Query(common.Filter)

	total, applications, err := a.applicationCtl.ListUserApplication(c, filter, &q.Query{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	})
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: applications,
	})
}
