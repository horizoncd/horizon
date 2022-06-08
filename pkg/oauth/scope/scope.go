package scope

import (
	"g.hz.netease.com/horizon/pkg/config/oauth"
	"g.hz.netease.com/horizon/pkg/rbac/types"
)

type Service interface {
	GetRulesByScope([]string) []types.Role
	GetAllScopeNames([]string) []string
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

func (f *fileScopeService) GetAllScopeNames([]string) []string {
	var scopeNames = make([]string, 0)
	for _, role := range f.Roles {
		scopeNames = append(scopeNames, role.Name)
	}
	return scopeNames
}
