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

package environmentregion

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/environmentregion"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
)

const (
	// param
	_environmentRegionIDParam = "environmentRegionID"
	_environmentNameQuery     = "environmentName"
)

type API struct {
	environmentRegionCtl environmentregion.Controller
}

func NewAPI(ctl environmentregion.Controller) *API {
	return &API{environmentRegionCtl: ctl}
}

func (a *API) List(c *gin.Context) {
	env := c.Query(_environmentNameQuery)
	var envRegions environmentregion.EnvironmentRegions
	var err error
	if env == "" {
		envRegions, err = a.environmentRegionCtl.ListAll(c)
	} else {
		envRegions, err = a.environmentRegionCtl.ListByEnvironment(c, env)
	}

	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg("failed to list environmentRegions"))
		return
	}

	response.SuccessWithData(c, envRegions)
}

func (a *API) Create(c *gin.Context) {
	var request *environmentregion.CreateEnvironmentRegionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	id, err := a.environmentRegionCtl.CreateEnvironmentRegion(c, request)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) SetDefault(c *gin.Context) {
	idStr := c.Param(_environmentRegionIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	err = a.environmentRegionCtl.SetEnvironmentRegionToDefault(c, uint(id))
	if err != nil {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
	}

	response.Success(c)
}

func (a *API) DeleteByID(c *gin.Context) {
	idStr := c.Param(_environmentRegionIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	err = a.environmentRegionCtl.DeleteByID(c, uint(id))
	if err != nil {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
	}

	response.Success(c)
}

func (a *API) SetIfAutoFree(c *gin.Context) {
	idStr := c.Param(_environmentRegionIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid id: %s, err: %s",
			idStr, err.Error())))
		return
	}

	var request environmentregion.SetAutoFreeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	err = a.environmentRegionCtl.SetEnvironmentRegionIfAutoFree(c, uint(id), request.AutoFree)
	if err != nil {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
	}

	response.Success(c)
}
