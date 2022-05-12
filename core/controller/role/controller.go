package role

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/rbac/types"
	"g.hz.netease.com/horizon/pkg/util/errors"
)

type Controller interface {
	ListRole(ctx context.Context) ([]types.Role, error)
}

func NewController(service role.Service) Controller {
	return &controller{roleService: service}
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
