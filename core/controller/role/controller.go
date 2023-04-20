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

package role

import (
	"context"
	"net/http"

	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/rbac/types"
	"github.com/horizoncd/horizon/pkg/util/errors"
)

type Controller interface {
	ListRole(ctx context.Context) ([]types.Role, error)
}

func NewController(param *param.Param) Controller {
	return &controller{roleService: param.RoleService}
}

type controller struct {
	roleService role.Service
}

func (c controller) ListRole(ctx context.Context) ([]types.Role, error) {
	const op = "role *controller: list role"
	roles, err := c.roleService.ListRole(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError, err.Error())
	}
	return roles, nil
}
