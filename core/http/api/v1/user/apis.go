package user

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/user"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

// path variable
const (
	_userIDParam = "userID"
)

type API struct {
	userCtl user.Controller
}

func NewAPI(ctl user.Controller) *API {
	return &API{
		userCtl: ctl,
	}
}

func (a *API) List(c *gin.Context) {
	queryName := c.Query(common.UserQueryName)
	keywords := q.KeyWords{}
	if queryName != "" {
		keywords[common.UserQueryName] = queryName
	}

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
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsgf(
			"failed to update user:\n"+
				"id = %v\nerr = %v", userID, err))
		return
	}

	response.SuccessWithData(c, updatedUser)
}
