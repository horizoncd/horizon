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

package build

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/build"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type API struct {
	buildCtl build.Controller
}

func NewAPI(buildCtl build.Controller) *API {
	return &API{buildCtl: buildCtl}
}

func (a *API) Get(c *gin.Context) {
	const op = "build: schemaGet"
	schema, err := a.buildCtl.GetSchema(c)
	if err != nil {
		log.WithFiled(c, "op", op).Errorf(err.Error())
		response.AbortWithRequestError(c, common.InternalError, err.Error())
	}
	response.SuccessWithData(c, schema)
}
