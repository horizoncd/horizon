package idp

import (
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/idp"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	usermodel "g.hz.netease.com/horizon/pkg/user/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var (
	_oidcCode    = "code"
	_oidcState   = "state"
	_redirectURL = "redirectUrl"
)

type API struct {
	idpCtrl idp.Controller
	store   sessions.Store
}

func NewAPI(idpCtrl idp.Controller, store sessions.Store) *API {
	return &API{
		idpCtrl: idpCtrl,
		store:   store,
	}
}

func (a *API) ListAuthEndpoints(c *gin.Context) {
	redirectURL := c.Query(_redirectURL)
	endpoints, err := a.idpCtrl.ListAuthEndpoints(c, redirectURL)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	c.PureJSON(http.StatusOK, response.NewResponseWithData(endpoints))
}

func (a *API) ListIDPs(c *gin.Context) {
	idps, err := a.idpCtrl.ListIDPs(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, idps)
}

func (a *API) LoginCallback(c *gin.Context) {
	code := c.Query(_oidcCode)
	state := c.Query(_oidcState)
	if code == "" || state == "" {
		response.AbortWithRPCError(c,
			rpcerror.ParamError.WithErrMsgf(
				"code and state should not be empty:\n"+
					"code = %s\n state = %s", code, state))
		return
	}

	var (
		user *usermodel.User
		err  error
	)

	if user, err = a.idpCtrl.Login(c, code, state); err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	session := a.getSession(c)
	if session == nil {
		return
	}

	session.Values[common.SessionKeyAuthUser] = &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	}

	if err := session.Save(c.Request, c.Writer); err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf(
				"saving session into backend or response failed:\n"+
					"err = %v", err))
		return
	}

	response.Success(c)
}

func (a *API) Logout(c *gin.Context) {
	session := a.getSession(c)
	if session == nil {
		return
	}

	session.Options.MaxAge = -1

	if err := session.Save(c.Request, c.Writer); err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf(
				"saving session into backend or response failed:\n"+
					"err = %v", err))
		return
	}
}

func (a *API) getSession(c *gin.Context) *sessions.Session {
	session, err := a.store.Get(c.Request, common.CookieKeyAuth)
	if err != nil {
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf(
				"session not found:\n"+
					"session name = %s\n err = %v", common.CookieKeyAuth, err))
		return nil
	}
	return session
}
