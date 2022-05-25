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
	DefaultScopeRole string
	Roles            []types.Role
}

func NewFileScopeService(config oauth.Config) Service {
	return &fileScopeService{
		DefaultScopeRole: config.DefaultScopeRole,
		Roles:            config.Roles,
	}
}

var _ Service = &fileScopeService{}

func (f *fileScopeService) GetRulesByScope(scopes []string) []types.Role {
	var roles = make([]types.Role, 0)
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
