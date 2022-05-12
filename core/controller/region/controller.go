package region

import (
	"context"

	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/region/models"
)

type Controller interface {
	ListRegions(ctx context.Context) ([]*models.RegionEntity, error)
}

func NewController() Controller {
	return &controller{regionMgr: regionmanager.Mgr}
}

type controller struct {
	regionMgr regionmanager.Manager
}

func (c controller) ListRegions(ctx context.Context) ([]*models.RegionEntity, error) {
	return c.regionMgr.ListRegionEntities(ctx)
}
