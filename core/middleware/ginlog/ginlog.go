package ginlog

import (
	"fmt"
	"io"
	"time"

	"g.hz.netease.com/horizon/core/middleware/requestid"
	"github.com/gin-gonic/gin"
)

func Middleware(output io.Writer, skipPaths ...string) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if params.IsOutputColor() {
				statusColor = params.StatusCodeColor()
				methodColor = params.MethodColor()
				resetColor = params.ResetColor()
			}
			if params.Latency > time.Minute {
				// Truncate in a golang < 1.8 safe way
				params.Latency = params.Latency - params.Latency%time.Second
			}
			rid := ""
			if v, ok := params.Keys[requestid.HeaderXRequestID].(string); ok {
				rid = v
			}
			return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s | %s |%s %-7s %s %#v\n%s",
				params.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, params.StatusCode, resetColor,
				params.Latency,
				params.ClientIP,
				rid,
				methodColor, params.Method, resetColor,
				params.Path,
				params.ErrorMessage,
			)
		},
		Output:    output,
		SkipPaths: skipPaths,
	})
}
