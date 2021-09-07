package requestid

import (
	"g.hz.netease.com/horizon/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

// HeaderXRequestID X-Request-ID header
const HeaderXRequestID = "X-Request-ID"

// Middleware add X-Request-ID header in the http request when not exist
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		rid := c.Request.Header.Get(HeaderXRequestID)
		if rid == "" {
			rid = uuid.NewV4().String()
		}
		c.Set(HeaderXRequestID, rid)
		c.Header(HeaderXRequestID, rid)
		c.Next()
	}, skippers...)
}
