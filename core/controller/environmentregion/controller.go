package environmentregion

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environmentregion/manager"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	ListRegionsByEnvironment(ctx context.Context, environment string) (Regions, error)
	CreateEnvironmentRegion(ctx context.Context, request *CreateEnvironmentRegionRequest) (uint, error)
	UpdateEnvironmentRegionByID(ctx context.Context, id uint, request *UpdateEnvironmentRegionRequest) error
}

var _ Controller = (*controller)(nil)

func NewController() Controller {
	return &controller{
		envRegionMgr: manager.Mgr,
	}
}

type controller struct {
	envRegionMgr manager.Manager
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

func (c *controller) UpdateEnvironmentRegionByID(ctx context.Context, id uint,
	request *UpdateEnvironmentRegionRequest) error {
	err := c.envRegionMgr.UpdateEnvironmentRegionByID(ctx, id, &models.EnvironmentRegion{
		EnvironmentName: request.EnvironmentName,
		RegionName:      request.RegionName,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) ListRegionsByEnvironment(ctx context.Context, environment string) (_ Regions, err error) {
	regions, err := c.envRegionMgr.ListRegionsByEnvironment(ctx, environment)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions), nil
}
