package application

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/application"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_extraOwner = "extraOwner"
)

type API struct {
	applicationCtl application.Controller
}

func NewAPI(applicationCtl application.Controller) *API {
	return &API{
		applicationCtl: applicationCtl,
	}
}

func (a *API) Create(c *gin.Context) {
	const op = "application: create"
	groupIDStr := c.Param(common.ParamGroupID)
	groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid groupID: %s, err: %s",
			groupIDStr, err.Error())))
		return
	}
	var request *application.CreateOrUpdateApplicationRequestV2
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	extraOwners := c.QueryArray(_extraOwner)
	if request.ExtraMembers == nil {
		request.ExtraMembers = make(map[string]string)
	}
	for _, roleOfMember := range request.ExtraMembers {
		if !role.CheckRoleIfValid(roleOfMember) {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("extra member is invalid"))
			return
		}
	}
	for _, owner := range extraOwners {
		request.ExtraMembers[owner] = role.Owner
	}

	resp, err := a.applicationCtl.CreateApplicationV2(c, uint(groupID), request)
	if err != nil {
		if perror.Cause(err) == herrors.ErrNameConflict {
			response.AbortWithRPCError(c, rpcerror.ConflictError.WithErrMsg(err.Error()))
			return
		} else if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithDataV2(c, resp)
}

func (a *API) Update(c *gin.Context) {
	const op = "application: update v2"
	var request *application.CreateOrUpdateApplicationRequestV2
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	err = a.applicationCtl.UpdateApplicationV2(c, uint(appID), request)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ApplicationInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		} else if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) Get(c *gin.Context) {
	op := "application: get v2"
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
		return
	}
	var resp *application.GetApplicationResponseV2
	if resp, err = a.applicationCtl.GetApplicationV2(c, uint(appID)); err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.ApplicationInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		}

		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithDataV2(c, resp)
}
