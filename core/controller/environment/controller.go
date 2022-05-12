package environment

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environment/manager"
	"g.hz.netease.com/horizon/pkg/environment/models"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error
	ListEnvironments(ctx context.Context) (Environments, error)
	ListRegionsByEnvironment(ctx context.Context, environment string) (Regions, error)
}

var _ Controller = (*controller)(nil)

func NewController() Controller {
	return &controller{
		envMgr: manager.Mgr,
	}
}

type controller struct {
	envMgr manager.Manager
}

func (c *controller) UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error {
	return c.envMgr.UpdateByID(ctx, id, &models.Environment{
		DisplayName:   request.DisplayName,
		DefaultRegion: request.DefaultRegion,
	})
}

func (c *controller) ListEnvironments(ctx context.Context) (_ Environments, err error) {
	envs, err := c.envMgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	return ofEnvironmentModels(envs), nil
}

func (c *controller) ListRegionsByEnvironment(ctx context.Context, environment string) (_ Regions, err error) {
	regions, err := c.envMgr.ListRegionsByEnvironment(ctx, environment)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions), nil
}
