package environment

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environment/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	ListEnvironments(ctx context.Context) (Environments, error)
	ListRegionsByEnvironment(ctx context.Context, environment string) (Regions, error)
}

func NewController() Controller {
	return &controller{
		envMgr: manager.Mgr,
	}
}

type controller struct {
	envMgr manager.Manager
}

func (c *controller) ListEnvironments(ctx context.Context) (_ Environments, err error) {
	const op = "environment controller: list environments"
	defer wlog.Start(ctx, op).StopPrint()

	envs, err := c.envMgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofEnvironmentModels(envs), nil
}

func (c *controller) ListRegionsByEnvironment(ctx context.Context, environment string) (_ Regions, err error) {
	const op = "environment controller: list regions by environment"
	defer wlog.Start(ctx, op).StopPrint()

	regions, err := c.envMgr.ListRegionsByEnvironment(ctx, environment)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofRegionModels(regions), nil
}
