package environment

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	herror "g.hz.netease.com/horizon/core/errors"
	environmentmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	"g.hz.netease.com/horizon/pkg/environment/models"
	envregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

type Controller interface {
	Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error)
	UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error
	ListEnvironments(ctx context.Context) (Environments, error)
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Environment, error)
	GetByName(ctx context.Context, name string) (*Environment, error)
	// ListEnabledRegionsByEnvironment deprecated, will be removed later. list regions by the environment that are enabled
	ListEnabledRegionsByEnvironment(ctx context.Context, environment string) (regionmodels.RegionParts, error)
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		envMgr:       param.EnvMgr,
		envRegionMgr: param.EnvRegionMgr,
		regionMgr:    param.RegionMgr,
	}
}

type controller struct {
	envMgr       environmentmanager.Manager
	envRegionMgr envregionmanager.Manager
	regionMgr    regionmanager.Manager
}

func (c *controller) GetByID(ctx context.Context, id uint) (*Environment, error) {
	environment, err := c.envMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofEnvironmentModel(environment), nil
}

func (c *controller) GetByName(ctx context.Context, name string) (*Environment, error) {
	environment, err := c.envMgr.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return ofEnvironmentModel(environment), nil
}

func (c *controller) Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error) {
	environment, err := c.envMgr.CreateEnvironment(ctx, &models.Environment{
		Name:        request.Name,
		DisplayName: request.DisplayName,
		AutoFree:    request.AutoFree,
	})
	if err != nil {
		return 0, err
	}
	return environment.ID, nil
}

func (c *controller) UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error {
	// online environment is not allowed to open auto-free
	environment, err := c.envMgr.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if environment.Name == common.OnlineEnv && request.AutoFree {
		return perror.Wrap(herror.ErrEnvironmentUpdateInvalid,
			"online environment is not allowed to open auto-free")
	}
	return c.envMgr.UpdateByID(ctx, id, &models.Environment{
		DisplayName: request.DisplayName,
		AutoFree:    request.AutoFree,
	})
}

func (c *controller) ListEnvironments(ctx context.Context) (_ Environments, err error) {
	envs, err := c.envMgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	return ofEnvironmentModels(envs), nil
}

func (c *controller) ListEnabledRegionsByEnvironment(ctx context.Context, environment string) (
	regionmodels.RegionParts, error) {
	return c.envRegionMgr.ListEnabledRegionsByEnvironment(ctx, environment)
}

func (c *controller) DeleteByID(ctx context.Context, id uint) error {
	return c.envMgr.DeleteByID(ctx, id)
}
