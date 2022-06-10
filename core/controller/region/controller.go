package region

import (
	"context"

	"g.hz.netease.com/horizon/pkg/param"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/region/models"
)

type Controller interface {
	ListRegions(ctx context.Context) ([]*Region, error)
	Create(ctx context.Context, request *CreateRegionRequest) (uint, error)
	UpdateByID(ctx context.Context, id uint, request *UpdateRegionRequest) error
	DeleteByID(ctx context.Context, id uint) error
	GetByName(ctx context.Context, name string) (*Region, error)
	GetByID(ctx context.Context, id uint) (*Region, error)
}

func NewController(param *param.Param) Controller {
	return &controller{regionMgr: param.RegionMgr}
}

type controller struct {
	regionMgr regionmanager.Manager
}

func (c controller) GetByID(ctx context.Context, id uint) (*Region, error) {
	region, err := c.regionMgr.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ofRegionEntity(region), nil
}

func (c controller) GetByName(ctx context.Context, name string) (*Region, error) {
	entity, err := c.regionMgr.GetRegionEntity(ctx, name)
	if err != nil {
		return nil, err
	}

	return ofRegionEntity(entity), nil
}

func (c controller) DeleteByID(ctx context.Context, id uint) error {
	return c.regionMgr.DeleteByID(ctx, id)
}

func (c controller) UpdateByID(ctx context.Context, id uint, request *UpdateRegionRequest) error {
	err := c.regionMgr.UpdateByID(ctx, id, &models.Region{
		DisplayName:   request.DisplayName,
		Server:        request.Server,
		Certificate:   request.Certificate,
		IngressDomain: request.IngressDomain,
		HarborID:      request.HarborID,
		Disabled:      request.Disabled,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c controller) Create(ctx context.Context, request *CreateRegionRequest) (uint, error) {
	create, err := c.regionMgr.Create(ctx, &models.Region{
		Name:          request.Name,
		DisplayName:   request.DisplayName,
		Server:        request.Server,
		Certificate:   request.Certificate,
		IngressDomain: request.IngressDomain,
		HarborID:      request.HarborID,
	})
	if err != nil {
		return 0, err
	}

	return create.ID, nil
}

func (c controller) ListRegions(ctx context.Context) ([]*Region, error) {
	entities, err := c.regionMgr.ListRegionEntities(ctx)
	if err != nil {
		return nil, err
	}
	return ofRegionEntities(entities), nil
}
