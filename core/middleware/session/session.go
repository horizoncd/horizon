package session

import (
	"net/http"

	"g.hz.netease.com/horizon/core/pkg/session"
	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const SessionIDKey = "sessionid"

func Middleware(session session.Interface, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		sessionID, err := c.Cookie(SessionIDKey)
		if err != nil {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
			return
		}
		s, err := session.GetSession(c, sessionID)
		if err != nil {
			response.AbortWithInternalError(c, "GetSessionFailed", err.Error())
			return
		}
		if s == nil {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
			return
		}
	}, skippers...)
}
