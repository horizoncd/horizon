package login

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	sessionmiddleware "g.hz.netease.com/horizon/core/middleware/session"
	"g.hz.netease.com/horizon/core/pkg/session"
	"g.hz.netease.com/horizon/server/response"
	"github.com/satori/go.uuid"
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
	if err := controller.sessionManager.SetSession(sessionID, s); err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("failed to set session: %v", err))
	}

	// 写客户端cookie，记录当前用户登录的stateID
	c.SetCookie(sessionmiddleware.SessionIDKey, sessionID,
		60*60*30, "", "", false, false)

	response.SuccessWithData(c, controller.oidc.GetRedirectURL(c.Request.Host, sessionID))
}

// Callback - user login callback
func (controller *Controller) Callback(c *gin.Context) {
	code := c.Query("code")
	sessionID := c.Query("state")
	user, err := controller.oidc.GetUser(c.Request.Host, code)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("failed to get user from oidc: %v", err))
		return
	}
	session, err := controller.sessionManager.GetSession(sessionID)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("failed to get session: %v", err))
		return
	}
	if session == nil {
		response.AbortWithRequestError(c, fmt.Sprintf("session with state: %v not found", sessionID))
		return
	}
	session.User = user
	if err := controller.sessionManager.SetSession(sessionID, session); err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("failed to set session: %v", err))
	}

	redirectTo := "http://" + session.FromHost + session.RedirectURL
	log.Printf("redirectTo: %v", redirectTo)
	c.Redirect(http.StatusMovedPermanently, redirectTo)
}

// Logout - user logout
func (controller *Controller) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionmiddleware.SessionIDKey)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get cookie failed: %v", err))
		return
	}
	if err := controller.sessionManager.DeleteSession(sessionID); err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("delete session failed: %v", err))
		return
	}

	c.SetCookie(sessionmiddleware.SessionIDKey, "", -1, "", "", false, false)
	response.Success(c)
}

// UserStatus -
func (controller *Controller) UserStatus(c *gin.Context) {
	sessionID, err := c.Cookie(sessionmiddleware.SessionIDKey)
	if err != nil {
		response.AbortWithInternalError(c, fmt.Sprintf("get cookie failed: %v", err))
		return
	}
	session, err := controller.sessionManager.GetSession(sessionID)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	if session == nil || session.User == nil {
		response.Abort(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	u := session.User
	response.SuccessWithData(c, User{
		Name:  u.Name,
	})
}
