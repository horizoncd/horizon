package oauthapp

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/controller/oauthapp"
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
	oauthAppClientIDStr := c.Param(_oauthAppClientIDIDParam)
	oauthApp, err := a.oauthAppController.Get(c, oauthAppClientIDStr)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, oauthApp)
}
func (a *API) DeleteOauthApp(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDIDParam)
	err := a.oauthAppController.Delete(c, oauthAppClientIDStr)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) UpdateOauthApp(c *gin.Context) {
	oauthAppClientIDStr := c.Param(_oauthAppClientIDIDParam)
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
