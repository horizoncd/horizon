package role

import "g.hz.netease.com/horizon/pkg/rbac/types"

type Config struct {
	RolePriorityRankDesc []string     `yaml:"RolePriorityRankDesc"`
	Roles                []types.Role `yaml:"Roles"`
}
