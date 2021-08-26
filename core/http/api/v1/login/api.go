package login

import (
	"fmt"
	"net/http"

	sessionmiddleware "g.hz.netease.com/horizon/core/middleware/session"
	"g.hz.netease.com/horizon/core/pkg/session"
	"g.hz.netease.com/horizon/pkg/log"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

const(
	SetSessionFailed = "SetSessionFailed"
	GetSessionFailed = "GetSessionFailed"
	DeleteSessionFailed = "DeleteSessionFailed"
	SessionNotFound = "SessionNotFound"
	GetOIDCUserFailed    = "GetOIDCUserFailed"
	GetCookieFailed = "GetCookieFailed"
	Unauthorized = "Unauthorized"
)

// Login - user login
func (controller *Controller) Login(c *gin.Context) {
	fromHostParam := c.Query("fromHost")
	redirectURLParam := c.Query("redirectUrl")
	s := &session.Session{
		FromHost:    fromHostParam,
		RedirectURL: redirectURLParam,
	}
	sessionID := uuid.NewV4().String()
	if err := controller.sessionManager.SetSession(c, sessionID, s); err != nil {
		response.AbortWithInternalError(c, SetSessionFailed, fmt.Sprintf("failed to set session: %v", err))
	}

	// 写客户端cookie，记录当前用户登录的stateID
	c.SetCookie(sessionmiddleware.SessionIDKey, sessionID,
		60*60*30, "", "", false, false)

	response.SuccessWithData(c, controller.oidc.GetRedirectURL(c, c.Request.Host, sessionID))
}

// Callback - user login callback
func (controller *Controller) Callback(c *gin.Context) {
	logger := log.GetLogger(c)
	code := c.Query("code")
	sessionID := c.Query("state")
	user, err := controller.oidc.GetUser(c, c.Request.Host, code)
	if err != nil {
		response.AbortWithInternalError(c, GetOIDCUserFailed, fmt.Sprintf("failed to get user from oidc: %v", err))
		return
	}
	session, err := controller.sessionManager.GetSession(c, sessionID)
	if err != nil {
		response.AbortWithInternalError(c, GetSessionFailed, fmt.Sprintf("failed to get session: %v", err))
		return
	}
	if session == nil {
		response.AbortWithRequestError(c, SessionNotFound, fmt.Sprintf("session with state: %v not found", sessionID))
		return
	}
	session.User = user
	if err := controller.sessionManager.SetSession(c, sessionID, session); err != nil {
		response.AbortWithInternalError(c, SetSessionFailed, fmt.Sprintf("failed to set session: %v", err))
	}

	redirectTo := "http://" + session.FromHost + session.RedirectURL
	logger.Infof("redirectTo: %v", redirectTo)
	c.Redirect(http.StatusMovedPermanently, redirectTo)
}

// Logout - user logout
func (controller *Controller) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionmiddleware.SessionIDKey)
	if err != nil {
		response.AbortWithInternalError(c, GetCookieFailed, fmt.Sprintf("get cookie failed: %v", err))
		return
	}
	if err := controller.sessionManager.DeleteSession(c, sessionID); err != nil {
		response.AbortWithInternalError(c, DeleteSessionFailed, fmt.Sprintf("delete session failed: %v", err))
		return
	}

	c.SetCookie(sessionmiddleware.SessionIDKey, "", -1, "", "", false, false)
	response.Success(c)
}

// UserStatus -
func (controller *Controller) UserStatus(c *gin.Context) {
	sessionID, err := c.Cookie(sessionmiddleware.SessionIDKey)
	if err != nil {
		response.AbortWithInternalError(c, GetCookieFailed, fmt.Sprintf("get cookie failed: %v", err))
		return
	}
	session, err := controller.sessionManager.GetSession(c, sessionID)
	if err != nil {
		response.AbortWithInternalError(c, GetSessionFailed, err.Error())
		return
	}
	if session == nil || session.User == nil {
		response.Abort(c, http.StatusUnauthorized, Unauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	u := session.User
	response.SuccessWithData(c, User{
		Name: u.Name,
	})
}
