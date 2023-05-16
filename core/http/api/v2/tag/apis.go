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

package tag

import (
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/controller/tag"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

const (
	_resourceTypeParam = "resourceType"
	_resourceIDParam   = "resourceID"
	_tagKey            = "key"
)

type API struct {
	tagCtl tag.Controller
}

func NewAPI(tagCtl tag.Controller) *API {
	return &API{
		tagCtl: tagCtl,
	}
}

func (a *API) List(c *gin.Context) {
	const op = "tag: list"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	resp, err := a.tagCtl.List(c, resourceType, uint(resourceID))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) Update(c *gin.Context) {
	const op = "tag: update"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	var request *tag.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid request body, err: %s", err.Error())))
		return
	}
	err = a.tagCtl.Update(c, resourceType, uint(resourceID), request)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) ListSubResourceTags(c *gin.Context) {
	const op = "tag: list sub resource tags"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsg(fmt.Sprintf("invalid resource id: %s", resourceIDStr)))
		return
	}

	resp, err := a.tagCtl.ListSubResourceTags(c, resourceType, uint(resourceID))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetMetatagKeys(c *gin.Context) {
	keys, err := a.tagCtl.GetMetatagKeys(c)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, keys)
}

func (a *API) GetMetatagsByKey(c *gin.Context) {
	key := c.Query(_tagKey)
	metatags, err := a.tagCtl.GetMetatagsByKey(c, key)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, metatags)
}

func (a *API) CreateMetatags(c *gin.Context) {
	var createMetatagsRequest tag.CreateMetatagsRequest
	if err := c.ShouldBindJSON(&createMetatagsRequest); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
		return
	}
	if err := a.tagCtl.CreateMetatags(c, &createMetatagsRequest); err != nil {
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}
