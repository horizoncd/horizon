package environment

import (
	"context"

	environmentmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	"github.com/horizoncd/horizon/pkg/environment/models"
	"github.com/horizoncd/horizon/pkg/environment/service"
	envregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	"github.com/horizoncd/horizon/pkg/param"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
)

type Controller interface {
	Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error)
	UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error
	ListEnvironments(ctx context.Context) (Environments, error)
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Environment, error)
	GetByName(ctx context.Context, name string) (*Environment, error)
	// ListEnabledRegionsByEnvironment will be removed later. list regions by the environment that are enabled
	// Deprecated
	ListEnabledRegionsByEnvironment(ctx context.Context, environment string) (regionmodels.RegionParts, error)
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		autoFreeSvc:  param.AutoFreeSvc,
		envMgr:       param.EnvMgr,
		envRegionMgr: param.EnvRegionMgr,
		regionMgr:    param.RegionMgr,
	}
}

type controller struct {
	envMgr       environmentmanager.Manager
	envRegionMgr envregionmanager.Manager
	regionMgr    regionmanager.Manager
	autoFreeSvc  *service.AutoFreeSVC
}

func (c *controller) GetByID(ctx context.Context, id uint) (*Environment, error) {
	environment, err := c.envMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofEnvironmentModel(environment, c.autoFreeSvc.WhetherSupported(environment.Name)), nil
}

func (c *controller) GetByName(ctx context.Context, name string) (*Environment, error) {
	environment, err := c.envMgr.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return ofEnvironmentModel(environment, c.autoFreeSvc.WhetherSupported(environment.Name)), nil
}

func (c *controller) Create(ctx context.Context, request *CreateEnvironmentRequest) (uint, error) {
	environment, err := c.envMgr.CreateEnvironment(ctx, &models.Environment{
		Name:        request.Name,
		DisplayName: request.DisplayName,
	})
	if err != nil {
		return 0, err
	}
	return environment.ID, nil
}

func (c *controller) UpdateByID(ctx context.Context, id uint, request *UpdateEnvironmentRequest) error {
	return c.envMgr.UpdateByID(ctx, id, &models.Environment{
		DisplayName: request.DisplayName,
	})
}

func (c *controller) ListEnvironments(ctx context.Context) (_ Environments, err error) {
	envs, err := c.envMgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	return ofEnvironmentModels(envs, c.autoFreeSvc), nil
}

func (c *controller) ListEnabledRegionsByEnvironment(ctx context.Context, environment string) (
	regionmodels.RegionParts, error) {
	return c.envRegionMgr.ListEnabledRegionsByEnvironment(ctx, environment)
}

func (c *controller) DeleteByID(ctx context.Context, id uint) error {
	return c.envMgr.DeleteByID(ctx, id)
}
