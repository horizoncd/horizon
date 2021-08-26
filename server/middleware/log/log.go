package log

import (
	"g.hz.netease.com/horizon/pkg/log"
	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/middleware/requestid"
	"github.com/gin-gonic/gin"
)

// Middleware add logger to context
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		rid := c.Value(requestid.HeaderXRequestID)
		if rid != "" {
			logger := log.DefaultLogger()
			logger.Debugf("attach request id %s to the logger for the request %s %s",
				rid, c.Request.Method, c.Request.URL.Path)
			c.Set(log.LoggerKey(), logger.WithField("requestID", rid))
		}
		c.Next()
	}, skippers...)
}
