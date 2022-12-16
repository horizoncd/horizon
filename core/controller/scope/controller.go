package scope

import (
	"context"

	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	ListScopes(ctx context.Context) []BasicInfo
}

func NewController(param *param.Param) Controller {
	return &controller{scopeSvc: param.ScopeService}
}

type controller struct {
	scopeSvc scope.Service
}

func (c controller) ListScopes(ctx context.Context) []BasicInfo {
	var resp []BasicInfo
	scopes := c.scopeSvc.GetAllScopes()
	for _, scope := range scopes {
		resp = append(resp, BasicInfo{
			Name: scope.Name,
			Desc: scope.Desc,
		})
	}
	return resp
}
