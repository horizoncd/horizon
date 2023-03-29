package user

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/user"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/user/util"
	"github.com/horizoncd/horizon/pkg/util/log"
)

// path variable
const (
	_userIDParam = "userID"
	_linkIDParam = "linkID"
)

type API struct {
	userCtl user.Controller
	store   sessions.Store
}

func NewAPI(ctl user.Controller, store sessions.Store) *API {
	return &API{
		userCtl: ctl,
		store:   store,
	}
}

func (a *API) List(c *gin.Context) {
	queryName := c.Query(common.UserQueryName)
	keywords := q.KeyWords{}
	if queryName != "" {
		keywords[common.UserQueryName] = queryName
	}

	var userTypes []int
	userTypeStrs := c.QueryArray(common.UserQueryType)
	for _, userTypeStr := range userTypeStrs {
		userType, err := strconv.Atoi(userTypeStr)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsgf("invalid user type: %s, err: %v", userTypeStr, err))
			return
		}
		userTypes = append(userTypes, userType)
	}
	if len(userTypes) == 0 {
		userTypes = append(userTypes, usermodels.UserTypeCommon)
	}
	keywords[common.UserQueryType] = userTypes

	query := q.New(keywords).WithPagination(c)
	total, users, err := a.userCtl.List(c, query)
	if err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to get users: "+
				"err = %v", err))
		return
	}

	response.SuccessWithData(c,
		response.DataWithTotal{
			Total: total,
			Items: users,
		})
}

func (a *API) GetSelf(c *gin.Context) {
	userInCtx, err := common.UserFromContext(c)
	if err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("user in context not found: err = %v", err))
		return
	}
	userInDB, err := a.userCtl.GetByID(c, userInCtx.GetID())
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c,
				rpcerror.NotFoundError.WithErrMsgf("user not found: id = %v, err =  %v", userInCtx.GetID(), err))
			return
		}
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to get user by id: err =  %v", err))
		return
	}

	response.SuccessWithData(c, userInDB)
}

func (a *API) GetByID(c *gin.Context) {
	op := "user: get by id"
	uid := c.Param(_userIDParam)
	var (
		userID uint64
		err    error
	)

	if userID, err = strconv.ParseUint(uid, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("user ID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("userID not found or invalid"))
		return
	}

	userInDB, err := a.userCtl.GetByID(c, uint(userID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c,
				rpcerror.NotFoundError.WithErrMsgf("user not found: id = %v, err =  %v", userID, err))
			return
		}
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to get user by id: err =  %v", err))
		return
	}

	response.SuccessWithData(c, userInDB)
}

func (a *API) Update(c *gin.Context) {
	op := "user: update by id"
	uid := c.Param(_userIDParam)
	var (
		userID uint64
		err    error
	)

	if userID, err = strconv.ParseUint(uid, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("user ID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("userID not found or invalid"))
		return
	}

	params := user.UpdateUserRequest{}
	if err = c.ShouldBindJSON(&params); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	updatedUser, err := a.userCtl.UpdateByID(c, uint(userID), &params)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c,
				rpcerror.NotFoundError.WithErrMsgf("user not found: id = %v, err =  %v", userID, err))
			return
		}
		if err = perror.Cause(err); errors.Is(err, herrors.ErrForbidden) {
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsgf(
				"can not update user:\n"+
					"id = %v\nerr = %v", userID, err))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsgf(
			"failed to update user:\n"+
				"id = %v\nerr = %v", userID, err))
		return
	}

	response.SuccessWithData(c, updatedUser)
}

func (a *API) GetLinksByUser(c *gin.Context) {
	op := "user links: get links by user"
	uid := c.Param(_userIDParam)
	var (
		userID uint64
		err    error
	)

	if userID, err = strconv.ParseUint(uid, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("user ID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("userID not found or invalid"))
		return
	}

	links, err := a.userCtl.ListUserLinks(c, uint(userID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c,
				rpcerror.NotFoundError.WithErrMsgf("user not found: id = %v, err =  %v", userID, err))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsgf(
			"failed to list links:\n"+
				"id = %v\nerr = %v", userID, err))
		return
	}

	response.SuccessWithData(c, links)
}

func (a *API) DeleteLink(c *gin.Context) {
	op := "user links: delete user"
	idStr := c.Param(_linkIDParam)
	var (
		linkID uint64
		err    error
	)

	if linkID, err = strconv.ParseUint(idStr, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("user ID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("userID not found or invalid"))
		return
	}

	err = a.userCtl.DeleteLinksByID(c, uint(linkID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c,
				rpcerror.NotFoundError.WithErrMsgf("link not found: id = %v, err =  %v", linkID, err))
			return
		}
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsgf(
			"failed to delete link:\n"+
				"id = %v\nerr = %v", linkID, err))
		return
	}

	response.Success(c)
}

func (a *API) LoginWithPassword(c *gin.Context) {
	var request *user.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil ||
		request.Password == "" ||
		request.Email == "" {
		response.AbortWithRPCError(c,
			rpcerror.ParamError.WithErrMsg("request body is invalid"))
		return
	}

	user, err := a.userCtl.LoginWithPasswd(c, request)
	if err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsg(
				fmt.Sprintf("login failed, err: %v", err)))
		return
	}

	if user == nil {
		response.AbortWithRPCError(c,
			rpcerror.Unauthorized.WithErrMsg("login failed: email or password is incorrect!"))
		return
	}

	session, err := util.GetSession(a.store, c.Request)
	if err != nil {
		log.Errorf(c, "failed to get session: store = %#v, request = %#v, err = %v",
			a.store, c.Request, err)
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to get session: %v", err))
		return
	}

	err = util.SetSession(session, c.Request, c.Writer, user)
	if err != nil {
		log.Errorf(c, "failed to set session: store = %#v, request = %#v, err = %v",
			a.store, c.Request, err)
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to set session: %v", err))
		return
	}
}
