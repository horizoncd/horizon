package oauthapp

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/oauthapp"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
)

type API struct {
	oauthAppController oauthapp.Controller
}

func NewAPI(controller oauthapp.Controller) *API {
	return &API{
		oauthAppController: controller,
	}
}

func (a *API) CreateOauthApp(c *gin.Context) {
	groupIDStr := c.Param(_groupIDParam)
	groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid groupID: %s, err: %s",
			groupIDStr, err.Error())))
		return
	}
	var req *oauthapp.CreateOauthAPPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}
	resp, err := a.oauthAppController.Create(c, uint(groupID), *req)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) ListOauthApp(c *gin.Context) {
	groupIDStr := c.Param(_groupIDParam)
	groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid groupID: %s, err: %s",
			groupIDStr, err.Error())))
		return
	}
	apps, err := a.oauthAppController.List(c, uint(groupID))
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, apps)
}

func (a *API) GetOauthApp(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	oauthApp, err := a.oauthAppController.Get(c, oauthAppClientIDStr)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, oauthApp)
}
func (a *API) DeleteOauthApp(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	err := a.oauthAppController.Delete(c, oauthAppClientIDStr)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) UpdateOauthApp(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	var req *oauthapp.APPBasicInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}
	if oauthAppClientIDStr != req.ClientID {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("clientID mismatch"))
		return
	}
	oauthApp, err := a.oauthAppController.Update(c, *req)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, oauthApp)
}

func (a *API) CreateSecret(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	secret, err := a.oauthAppController.CreateSecret(c, oauthAppClientIDStr)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.OAuthInDB {
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
				return
			}
		}
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, secret)
}

func (a *API) ListSecret(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	secrets, err := a.oauthAppController.ListSecret(c, oauthAppClientIDStr)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.OAuthInDB {
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
				return
			}
		}
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, secrets)
}

func (a *API) DeleteSecret(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDParam)
	oauthClientSecretID := c.Param(_oauthClientSecretID)
	clientSecretID, err := strconv.ParseUint(oauthClientSecretID, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid SecretID: %s, err: %s",
			oauthClientSecretID, err.Error())))
		return
	}
	err = a.oauthAppController.DeleteSecret(c, oauthAppClientIDStr, uint(clientSecretID))
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if e.Source == herrors.OAuthInDB {
				response.AbortWithUnauthorized(c, common.Unauthorized, err.Error())
				return
			}
		}
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.Success(c)
}
