package role

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

	roleconfig "g.hz.netease.com/horizon/pkg/config/role"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"g.hz.netease.com/horizon/pkg/util/log"
	"gopkg.in/yaml.v2"
)

type CompResult uint8

const (
	RoleEqual CompResult = iota
	RoleBigger
	RoleSmaller
	RoleCanNotCompare
)

const (
	Owner string = "owner"
)

var (
	ErrorRoleNotFound   = errors.New("role not found")
	ErrorLoadCheckError = errors.New("load check error")
)

type Service interface {
	// ListRole List all the role
	ListRole(ctx context.Context) ([]types.Role, error)
	// GetRole get role by the role name
	GetRole(ctx context.Context, roleName string) (*types.Role, error)
	// RoleCompare compare if the role1's permission is higher than role2
	RoleCompare(ctx context.Context, role1, role2 string) (CompResult, error)
	// GetDefaultRole return the default role(no default role will return nil)
	GetDefaultRole(ctx context.Context) *types.Role
}

type roleRankMapItem struct {
	rank int
	role types.Role
}
type fileRoleService struct {
	RolePriorityRankDesc []string
	DefaultRoleName      string
	Roles                []types.Role

	DefaultRole *types.Role
	roleRankMap map[string]roleRankMapItem
}

func NewFileRoleFrom2(ctx context.Context, config roleconfig.Config) (Service, error) {
	fRole := fileRoleService{
		RolePriorityRankDesc: config.RolePriorityRankDesc,
		Roles:                config.Roles,
		DefaultRoleName:      config.DefaultRole,
	}
	// check
	if len(fRole.Roles) != len(fRole.RolePriorityRankDesc) {
		log.Error(ctx, "role number in RolePriorityRank not equal with Roles")
		return nil, ErrorLoadCheckError
	}

	roleRankMap := make(map[string]int)
	for i, roleStr := range fRole.RolePriorityRankDesc {
		roleRankMap[roleStr] = i
	}

	fRole.roleRankMap = make(map[string]roleRankMapItem)
	for _, role := range fRole.Roles {
		rankNum, ifOk := roleRankMap[role.Name]
		if !ifOk {
			log.Errorf(ctx, "RolePriorityRankDesc array doesn't contains role %s\n", role.Name)
			return nil, ErrorLoadCheckError
		}
		fRole.roleRankMap[role.Name] = roleRankMapItem{
			rank: rankNum,
			role: role,
		}
	}

	if fRole.DefaultRoleName != "" {
		defaultRole, ok := fRole.roleRankMap[fRole.DefaultRoleName]
		if !ok {
			log.WithFiled(ctx, "DefaultRole", fRole.DefaultRole).Error("not found")
			return nil, ErrorRoleNotFound
		}
		fRole.DefaultRole = &defaultRole.role
	}

	return &fRole, nil
}

func NewFileRole(ctx context.Context, reader io.Reader) (Service, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var config roleconfig.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}
	return NewFileRoleFrom2(ctx, config)
}

func (fRole *fileRoleService) ListRole(ctx context.Context) ([]types.Role, error) {
	return fRole.Roles, nil
}

func (fRole *fileRoleService) GetRole(ctx context.Context, roleName string) (*types.Role, error) {
	role, ifOk := fRole.roleRankMap[roleName]
	if !ifOk {
		return nil, ErrorRoleNotFound
	}
	return &role.role, nil
}

func (fRole *fileRoleService) GetDefaultRole(ctx context.Context) *types.Role {
	return fRole.DefaultRole
}

func (fRole *fileRoleService) RoleCompare(ctx context.Context, role1, role2 string) (CompResult, error) {
	item1, ifOk1 := fRole.roleRankMap[role1]
	item2, ifOk2 := fRole.roleRankMap[role2]

	if !ifOk1 || !ifOk2 {
		log.Errorf(ctx, "role %s cannot found", role1)
		return RoleCanNotCompare, ErrorRoleNotFound
	}
	rankRole1 := item1.rank
	rankRole2 := item2.rank
	if rankRole1 < rankRole2 {
		return RoleBigger, nil
	} else if rankRole1 > rankRole2 {
		return RoleSmaller, nil
	} else {
		return RoleEqual, nil
	}
}
