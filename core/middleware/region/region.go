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

package region

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/applicationregion"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

// http params for create cluster api
var _urlPattern = regexp.MustCompile(`/apis/core/v[12]/applications/(\d+)/clusters`)

const (
	_method     = http.MethodPost
	_scopeParam = "scope"
)

// Middleware to set region for create cluster API
func Middleware(param *param.Param, applicationRegionCtl applicationregion.Controller,
	skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// not create cluster api, skip
		if !_urlPattern.MatchString(c.Request.URL.Path) || c.Request.Method != _method {
			c.Next()
			return
		}
		matches := _urlPattern.FindStringSubmatch(c.Request.URL.Path)
		applicationIDStr := matches[1]
		applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("invalid applicationID: %v", applicationID))
			return
		}
		// scope format is: {environment}/{region}
		query := c.Request.URL.Query()
		scope := query.Get(_scopeParam)
		params := strings.Split(scope, "/")
		// invalid scope
		if scope == "" || len(params) > 2 {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid scope!")
			return
		}
		// params length is 2, satisfies {environment}/{region} format, skip
		if len(params) == 2 {
			c.Next()
			return
		}
		environment := params[0]

		applicationRegions, err := applicationRegionCtl.List(c, uint(applicationID))
		if err != nil {
			response.AbortWithInternalError(c,
				fmt.Sprintf("failed to get applicationRegions by id: %v", applicationID))
			return
		}
		r, ok := getRegion(c, applicationRegions, environment, param)
		if !ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(
				fmt.Sprintf("cannot find region for environment %v, application %v",
					environment, applicationID)))
			return
		}

		query.Set(_scopeParam, fmt.Sprintf("%v/%v", environment, r))
		c.Request.URL.RawQuery = query.Encode()
		c.Next()
	}, skippers...)
}

func getRegion(c *gin.Context, applicationRegions applicationregion.ApplicationRegion,
	environment string, p *param.Param) (string, bool) {
	for _, applicationRegion := range applicationRegions {
		if applicationRegion.Environment == environment {
			r := applicationRegion.Region
			region, err := p.RegionMgr.GetRegionByName(c, r)
			if err != nil {
				log.Errorf(c, "query region failed: %s, err: %+v", r, err)
				return "", false
			}
			if region.Disabled {
				log.Errorf(c, "region disabled: %s", r)
				return "", false
			}
			return r, true
		}
	}
	return "", false
}
