package log

import (
	"fmt"

	middleware2 "g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

// Middleware add logger to context
func Middleware(skippers ...middleware2.Skipper) gin.HandlerFunc {
	return middleware2.New(func(c *gin.Context) {
		rid := c.Value(requestid.HeaderXRequestID)
		if rid != "" {
			c.Set(log.Key(), fmt.Sprintf("[%v] ", rid))
		}
		c.Next()
	}, skippers...)
}
