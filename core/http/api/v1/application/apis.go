package application

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/application"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/request"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	// param
	_extraOwner = "extraOwner"
	_groupIDStr = "groupID"
	_envQuery   = "env"
	_hard       = "hard"
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
	op := "application: get"
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
		return
	}
	var res *application.GetApplicationResponse
	if res, err = a.applicationCtl.GetApplication(c, uint(appID)); err != nil {
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
	response.SuccessWithData(c, res)
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

	extraOwners := c.QueryArray(_extraOwner)

	var request *application.CreateApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

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
	resp, err := a.applicationCtl.CreateApplication(c, uint(groupID), request)
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
	response.SuccessWithData(c, resp)
}

func (a *API) Update(c *gin.Context) {
	const op = "application: update"
	var request *application.UpdateApplicationRequest
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
	resp, err := a.applicationCtl.UpdateApplication(c, uint(appID), request)
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
	response.SuccessWithData(c, resp)
}

func (a *API) Transfer(c *gin.Context) {
	const op = "application: transfer"
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
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
			if e.Source == herrors.GroupInDB || e.Source == herrors.ApplicationInDB {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
		} else if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.Success(c)
}

func (a *API) Delete(c *gin.Context) {
	const op = "application: delete"
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
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
	if err := a.applicationCtl.DeleteApplication(c, uint(appID), hard); err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.GroupInDB || e.Source == herrors.ApplicationInDB {
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

// SearchApplication search all applications
func (a *API) SearchApplication(c *gin.Context) {
	const op = "application: search application"
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid pageNumber or "+
			"pageSize, err: %s", err.Error())))
		return
	}

	filter := c.Query(common.Filter)

	total, applications, err := a.applicationCtl.ListApplication(c, filter, q.Query{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	})
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.GroupInDB || e.Source == herrors.ApplicationInDB {
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

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: applications,
	})
}

// SearchMyApplication search all applications that authorized to current user
func (a *API) SearchMyApplication(c *gin.Context) {
	const op = "application: search my application"
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid pageNumber or "+
			"pageSize, err: %s", err.Error())))
		return
	}

	filter := c.Query(common.Filter)

	total, applications, err := a.applicationCtl.ListUserApplication(c, filter, &q.Query{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	})
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.GroupInDB || e.Source == herrors.ApplicationInDB {
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

	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: applications,
	})
}

func (a *API) GetSelectableRegionsByEnv(c *gin.Context) {
	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
		return
	}

	env := c.Query(_envQuery)
	if env == "" {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("env in URL query parameters cannot be empty"))
		return
	}

	regionParts, err := a.applicationCtl.GetSelectableRegionsByEnv(c, uint(appID), env)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
	}

	response.SuccessWithData(c, regionParts)
}
