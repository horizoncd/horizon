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

package environment

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/environment"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	// param
	_environmentParam = "environment"
)

type API struct {
	envCtl environment.Controller
}

func NewAPI(ctl environment.Controller) *API {
	return &API{
		envCtl: ctl,
	}
}

func (a *API) ListEnvironments(c *gin.Context) {
	envs, err := a.envCtl.ListEnvironments(c)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, envs)
}

// ListEnabledRegionsByEnvironment deprecated, use GetSelectableRegionsByEnv in environment api
func (a *API) ListEnabledRegionsByEnvironment(c *gin.Context) {
	env := c.Param(_environmentParam)
	regions, err := a.envCtl.ListEnabledRegionsByEnvironment(c, env)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, regions)
}

func (a *API) Update(c *gin.Context) {
	const op = "environment: update"

	envIDStr := c.Param(_environmentParam)
	envID, err := strconv.ParseUint(envIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	var request *environment.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	err = a.envCtl.UpdateByID(c, uint(envID), request)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}

func (a *API) Create(c *gin.Context) {
	const op = "environment: create"

	var request *environment.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("invalid request body, err: %s",
			err.Error())))
		return
	}

	id, err := a.envCtl.Create(c, request)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, id)
}

func (a *API) Delete(c *gin.Context) {
	const op = "environment: delete"
	envIDStr := c.Param(_environmentParam)
	envID, err := strconv.ParseUint(envIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	err = a.envCtl.DeleteByID(c, uint(envID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.Success(c)
}
func (a *API) GetByID(c *gin.Context) {
	const op = "environment: get"
	envIDStr := c.Param(_environmentParam)
	envID, err := strconv.ParseUint(envIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}

	environmentEntity, err := a.envCtl.GetByID(c, uint(envID))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, environmentEntity)
}
