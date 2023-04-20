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

package scope

import (
	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/core/controller/scope"
	"github.com/horizoncd/horizon/pkg/server/response"
)

type API struct {
	scopeCtrl scope.Controller
}

func NewAPI(controller scope.Controller) *API {
	return &API{scopeCtrl: controller}
}

func (a *API) ListScopes(c *gin.Context) {
	scopes := a.scopeCtrl.ListScopes(c)
	response.SuccessWithData(c, scopes)
}
