package oauth

import "g.hz.netease.com/horizon/pkg/rbac/types"

type Config struct {
	DefaultScopes []string     `yaml:"DefaultScope"`
	Roles         []types.Role `yaml:"Roles"`
}
