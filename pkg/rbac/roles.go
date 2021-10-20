package rbac

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

	"g.hz.netease.com/horizon/pkg/util/log"
	"gopkg.in/yaml.v2"
)

type RoleCompResult uint8

const (
	RoleEqual RoleCompResult = iota
	RoleBigger
	RoleSmaller
	RoleCanNotCompare
)

var (
	ErrorRoleNotFound   = errors.New("role not found")
	ErrorLoadCheckError = errors.New("load check error")
)

type Service interface {
	ListRole(ctx context.Context) ([]Role, error)
	GetRole(ctx context.Context, roleName string) (*Role, error)
	RoleCompare(ctx context.Context, role1, role2 string) (RoleCompResult, error)
}

type roleRankMapItem struct {
	rank int
	role Role
}
type fileRoleService struct {
	RolePriorityRankDesc []string `yaml:"RolePriorityRankDesc"`
	Roles                []Role   `yaml:"Roles"`
	roleRankMap          map[string]roleRankMapItem
}

func NewFileRole(ctx context.Context, reader io.Reader) (Service, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var fRole fileRoleService
	if err := yaml.Unmarshal(content, &fRole); err != nil {
		return nil, err
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
	return &fRole, nil
}

func (fRole *fileRoleService) ListRole(ctx context.Context) ([]Role, error) {
	return fRole.Roles, nil
}

func (fRole *fileRoleService) GetRole(ctx context.Context, roleName string) (*Role, error) {
	role, ifOk := fRole.roleRankMap[roleName]
	if !ifOk {
		return nil, ErrorRoleNotFound
	}
	return &role.role, nil
}

func (fRole *fileRoleService) RoleCompare(ctx context.Context, role1, role2 string) (RoleCompResult, error) {
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
