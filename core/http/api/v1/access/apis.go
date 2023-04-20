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

package access

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/controller/access"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type API struct {
	accessCtl access.Controller
}

func NewAPI(c access.Controller) *API {
	return &API{
		accessCtl: c,
	}
}

func (a *API) AccessReview(c *gin.Context) {
	const op = "access: access review"

	var request *access.ReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("request body is invalid, err: %v", err)))
		return
	}

	if len(request.APIs) == 0 {
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("apis should not be empty"))
		return
	}

	reviewResp, err := a.accessCtl.Review(c, request.APIs)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}

	response.SuccessWithData(c, reviewResp)
}
