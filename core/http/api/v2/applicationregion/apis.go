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

package applicationregion

import (
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/applicationregion"
	"github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"

	"github.com/gin-gonic/gin"
)

type API struct {
	applicationRegionCtl applicationregion.Controller
}

func NewAPI(applicationRegionCtl applicationregion.Controller) *API {
	return &API{
		applicationRegionCtl: applicationRegionCtl,
	}
}

func (a *API) List(c *gin.Context) {
	applicationIDStr := c.Param(common.ParamApplicationID)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	var res applicationregion.ApplicationRegion
	if res, err = a.applicationRegionCtl.List(c, uint(applicationID)); err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, res)
}

func (a *API) Update(c *gin.Context) {
	applicationIDStr := c.Param(common.ParamApplicationID)
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	var request applicationregion.ApplicationRegion
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	if err := a.applicationRegionCtl.Update(c, uint(applicationID), request); err != nil {
		switch perror.Cause(err).(type) {
		case *errors.HorizonErrNotFound:
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		default:
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		}
		return
	}

	response.Success(c)
}
