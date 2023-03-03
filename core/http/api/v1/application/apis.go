package application

import (
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/application"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/request"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	// param
	_extraOwner = "extraOwner"
	_groupIDStr = "groupID"
	_envQuery   = "env"
	_hard       = "hard"
	_cluster    = "cluster"
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

// List search all applications that authorized to current user
func (a *API) List(c *gin.Context) {
	const op = "application: list application"

	keywords := make(map[string]interface{})

	filter := c.Query(common.ApplicationQueryName)
	if filter != "" {
		keywords[common.ApplicationQueryName] = filter
	}

	template := c.Query(common.ApplicationQueryByTemplate)
	if template != "" {
		keywords[common.ApplicationQueryByTemplate] = template
	}

	release := c.Query(common.ApplicationQueryByRelease)
	if release != "" {
		keywords[common.ApplicationQueryByRelease] = release
	}

	idStr := c.Query(common.ApplicationQueryByUser)
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("user id is not a number"))
			return
		}
		keywords[common.ApplicationQueryByUser] = uint(id)
	}

	query := q.New(keywords).WithPagination(c)

	applications, total, err := a.applicationCtl.List(c, query)
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

func (a *API) GetApplicationPipelineStats(c *gin.Context) {
	pageNumber, pageSize, err := request.GetPageParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	appIDStr := c.Param(common.ParamApplicationID)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid appID: %s, err: %s",
			appIDStr, err.Error())))
		return
	}

	cluster := c.Query(_cluster)
	pipelineStats, count, err := a.applicationCtl.GetApplicationPipelineStats(c, uint(appID), cluster, pageNumber,
		pageSize)
	if err != nil {
		log.Errorf(c, "Get application pipelineStats failed, error: %+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, response.DataWithTotal{
		Total: count,
		Items: pipelineStats,
	})
}
