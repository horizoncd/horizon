package log

import (
	middleware "g.hz.netease.com/horizon/core/middleware"
	"g.hz.netease.com/horizon/core/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

// Middleware add logger to context
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		rid := c.Value(requestid.HeaderXRequestID)
		if rid != "" {
			c.Set(log.Key(), rid)
		}
		c.Next()
	}, skippers...)
}
