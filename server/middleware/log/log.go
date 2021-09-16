package log

import (
	"fmt"

	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/middleware/requestid"
	"g.hz.netease.com/horizon/util/log"
	"github.com/gin-gonic/gin"
)

// Middleware add logger to context
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		rid := c.Value(requestid.HeaderXRequestID)
		if rid != "" {
			c.Set(log.Key(), fmt.Sprintf("[%v] ", rid))
		}
		c.Next()
	}, skippers...)
}
