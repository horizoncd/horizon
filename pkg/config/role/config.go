package role

import "github.com/horizoncd/horizon/pkg/rbac/types"

type Config struct {
	RolePriorityRankDesc []string     `yaml:"RolePriorityRankDesc"`
	DefaultRole          string       `yaml:"DefaultRole"`
	Roles                []types.Role `yaml:"Roles"`
}
