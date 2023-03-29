package requestid

import (
	"context"

	"github.com/gin-gonic/gin"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	middleware "github.com/horizoncd/horizon/pkg/core/middleware"
	uuid "github.com/satori/go.uuid"
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

func FromContext(ctx context.Context) (string, error) {
	rid, ok := ctx.Value(HeaderXRequestID).(string)
	if !ok {
		return "", herrors.ErrFailedToGetRequestID
	}
	return rid, nil
}
