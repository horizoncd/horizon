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
	Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error)
	UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error
	ListEnvironments(ctx context.Context) (Environments, error)
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

func (c *controller) Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error) {
	environment, err := c.envMgr.CreateEnvironment(ctx, &models.Environment{
		Name:          request.Name,
		DisplayName:   request.DisplayName,
		DefaultRegion: request.DefaultRegion,
	})
	if err != nil {
		return 0, err
	}
	return environment.ID, nil
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
