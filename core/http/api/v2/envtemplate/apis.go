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

package envtemplate

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/envtemplate"
	"github.com/horizoncd/horizon/pkg/server/response"
)

const (
	_applicationIDParam = "applicationID"
	_envParam           = "environment"
)

type API struct {
	envTemplateCtl envtemplate.Controller
}

func NewAPI(envTemplateCtl envtemplate.Controller) *API {
	return &API{
		envTemplateCtl: envTemplateCtl,
	}
}

func (a *API) Get(c *gin.Context) {
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	env := c.Query(_envParam)

	var res *envtemplate.GetEnvTemplateResponse
	if res, err = a.envTemplateCtl.GetEnvTemplate(c, uint(appID), env); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, res)
}

func (a *API) Update(c *gin.Context) {
	appIDStr := c.Param(_applicationIDParam)
	appID, err := strconv.ParseUint(appIDStr, 10, 0)
	if err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
		return
	}

	env := c.Query(_envParam)

	var r *envtemplate.UpdateEnvTemplateRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}

	if err = a.envTemplateCtl.UpdateEnvTemplateV2(c, uint(appID), env, r); err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.Success(c)
}
