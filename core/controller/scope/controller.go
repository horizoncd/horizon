package scope

import (
	"context"

	"g.hz.netease.com/horizon/pkg/oauth/scope"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/rbac/types"
)

type Controller interface {
	ListScopes(ctx context.Context) []types.Role
}

func NewController(param *param.Param) Controller {
	return &controller{scopeSvc: param.ScopeService}
}

type controller struct {
	scopeSvc scope.Service
}

func (c controller) ListScopes(ctx context.Context) []types.Role {
	return c.scopeSvc.GetAllScopes()
}
