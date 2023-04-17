package log

import (
	"github.com/gin-gonic/gin"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/pkg/util/log"
)

// Middleware add logger to context.
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		rid := c.Value(requestid.HeaderXRequestID)
		if rid != "" {
			c.Set(log.Key(), rid)
		}
		c.Next()
	}, skippers...)
}
