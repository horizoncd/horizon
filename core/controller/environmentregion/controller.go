package environmentregion

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environmentregion/manager"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	ListAll(ctx context.Context) (EnvironmentRegions, error)
	ListByEnvironment(ctx context.Context, environment string) (EnvironmentRegions, error)
	CreateEnvironmentRegion(ctx context.Context, request *CreateEnvironmentRegionRequest) (uint, error)
	SetEnvironmentRegionToDefault(ctx context.Context, id uint) error
	DeleteByID(ctx context.Context, id uint) error
}

var _ Controller = (*controller)(nil)

func NewController() Controller {
	return &controller{
		envRegionMgr: manager.Mgr,
		regionMgr:    regionmanager.Mgr,
	}
}

type controller struct {
	envRegionMgr manager.Manager
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
