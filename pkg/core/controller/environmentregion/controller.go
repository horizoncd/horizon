package environmentregion

import (
	"context"

	environmentregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	"github.com/horizoncd/horizon/pkg/environmentregion/models"
	"github.com/horizoncd/horizon/pkg/param"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
)

type Controller interface {
	ListAll(ctx context.Context) (EnvironmentRegions, error)
	ListByEnvironment(ctx context.Context, environment string) (EnvironmentRegions, error)
	CreateEnvironmentRegion(ctx context.Context, request *CreateEnvironmentRegionRequest) (uint, error)
	SetEnvironmentRegionToDefault(ctx context.Context, id uint) error
	DeleteByID(ctx context.Context, id uint) error
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		envRegionMgr: param.EnvRegionMgr,
		regionMgr:    param.RegionMgr,
	}
}

type controller struct {
	envRegionMgr environmentregionmanager.Manager
	regionMgr    regionmanager.Manager
}

func (c *controller) ListByEnvironment(ctx context.Context, environment string) (EnvironmentRegions, error) {
	environmentRegions, err := c.envRegionMgr.ListByEnvironment(ctx, environment)
	if err != nil {
		return nil, err
	}
	regions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions, environmentRegions), nil
}

// DeleteByID implements Controller
func (c *controller) DeleteByID(ctx context.Context, id uint) error {
	return c.envRegionMgr.DeleteByID(ctx, id)
}

func (c *controller) CreateEnvironmentRegion(ctx context.Context,
	request *CreateEnvironmentRegionRequest) (uint, error) {
	environmentRegion, err := c.envRegionMgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: request.EnvironmentName,
		RegionName:      request.RegionName,
	})
	if err != nil {
		return 0, err
	}
	return environmentRegion.ID, nil
}

func (c *controller) ListAll(ctx context.Context) (_ EnvironmentRegions, err error) {
	environmentRegions, err := c.envRegionMgr.ListAllEnvironmentRegions(ctx)
	if err != nil {
		return nil, err
	}
	regions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions, environmentRegions), nil
}

func (c *controller) SetEnvironmentRegionToDefault(ctx context.Context, id uint) error {
	return c.envRegionMgr.SetEnvironmentRegionToDefaultByID(ctx, id)
}
