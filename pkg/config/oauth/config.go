package oauth

import "g.hz.netease.com/horizon/pkg/rbac/types"

type Config struct {
	DefaultScopeRole string       `yaml:"DefaultScopeRole"`
	Roles            []types.Role `yaml:"Roles"`
}
