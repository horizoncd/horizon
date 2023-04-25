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

package requestid

import (
	"context"

	"github.com/gin-gonic/gin"
	herrors "github.com/horizoncd/horizon/core/errors"
	middleware "github.com/horizoncd/horizon/core/middleware"
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
