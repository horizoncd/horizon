package region

import (
	"context"

	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/region/models"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	ListRegions(ctx context.Context) ([]*models.RegionEntity, error)
	Create(ctx context.Context, region *CreateRegionRequest) (uint, error)
}

func NewController() Controller {
	return &controller{regionMgr: regionmanager.Mgr}
}

type controller struct {
	regionMgr regionmanager.Manager
}

func (c controller) Create(ctx context.Context, region *CreateRegionRequest) (uint, error) {
	create, err := c.regionMgr.Create(ctx, &models.Region{
		Name:          region.Name,
		DisplayName:   region.DisplayName,
		Server:        region.Server,
		Certificate:   region.Certificate,
		IngressDomain: region.IngressDomain,
		HarborID:      region.HarborID,
	})
	if err != nil {
		return 0, err
	}

	return create.ID, nil
}

func (c controller) ListRegions(ctx context.Context) ([]*models.RegionEntity, error) {
	return c.regionMgr.ListRegionEntities(ctx)
}
