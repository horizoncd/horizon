// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ginlog

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/middleware/requestid"
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
