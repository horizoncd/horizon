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
	"github.com/horizoncd/horizon/pkg/config/oauth"
	"github.com/horizoncd/horizon/pkg/rbac/types"
)

type Service interface {
	GetRulesByScope([]string) []types.Role
	GetAllScopeNames() []string
	GetAllScopes() []types.Role
}

type fileScopeService struct {
	DefaultScopes []string
	DefaultRoles  []types.Role
	Roles         []types.Role
}

func NewFileScopeService(config oauth.Scopes) (Service, error) {
	defaultRoles := make([]types.Role, 0)

	getRole := func(scopeName string) *types.Role {
		for _, role := range config.Roles {
			if role.Name == scopeName {
				return &role
			}
		}
		return nil
	}
	for _, scope := range config.DefaultScopes {
		role := getRole(scope)
		if role != nil {
			defaultRoles = append(defaultRoles, *role)
		}
	}
	return &fileScopeService{
		DefaultScopes: config.DefaultScopes,
		DefaultRoles:  defaultRoles,
		Roles:         config.Roles,
	}, nil
}

var _ Service = &fileScopeService{}

func (f *fileScopeService) GetRulesByScope(scopes []string) []types.Role {
	var roles = make([]types.Role, 0)
	if len(scopes) == 0 || (len(scopes) == 1 && scopes[0] == "") {
		return append(roles, f.DefaultRoles...)
	}
	for _, scope := range scopes {
		for _, role := range f.Roles {
			if role.Name == scope {
				roles = append(roles, role)
			}
		}
	}
	return roles
}

func (f *fileScopeService) GetAllScopeNames() []string {
	var scopeNames = make([]string, 0)
	for _, role := range f.Roles {
		scopeNames = append(scopeNames, role.Name)
	}
	return scopeNames
}

func (f *fileScopeService) GetAllScopes() []types.Role {
	return f.Roles
}
