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

package health

import (
	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/response"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine) {
	engine.GET("/health", healthCheck)
	engine.GET("/ready", readinessCheck)

	// authn && authz required, admin only
	engine.POST("/apis/core/v2/offline", offline)
	engine.POST("/apis/core/v2/online", online)
}

func healthCheck(c *gin.Context) {
	response.Success(c)
}

var ready = true

func online(c *gin.Context) {
	ready = true
	response.Success(c)
}

func offline(c *gin.Context) {
	ready = false
	response.Success(c)
}

func readinessCheck(c *gin.Context) {
	if ready {
		response.Success(c)
		return
	}
	response.AbortWithServiceUnavailable(c, common.ServiceUnavailable, "Service is offline")
}
