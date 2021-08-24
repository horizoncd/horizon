package session

import (
	"fmt"
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
			response.AbortWithInternalError(c, fmt.Sprintf("get cookie failed: %v", err))
			return
		}
		s, err := session.GetSession(sessionID)
		if err != nil{
			response.AbortWithInternalError(c, err.Error())
			return
		}
		if s == nil {
			response.Abort(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}
	}, skippers...)
}