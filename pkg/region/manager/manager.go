package manager

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
	"g.hz.netease.com/horizon/pkg/region/models"
	registrydao "g.hz.netease.com/horizon/pkg/registry/dao"
	registrymodels "g.hz.netease.com/horizon/pkg/registry/models"
	tagdao "g.hz.netease.com/horizon/pkg/tag/dao"
	"gorm.io/gorm"
)

type Manager interface {
	// Create a region
	Create(ctx context.Context, region *models.Region) (*models.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*models.Region, error)
	// ListRegionEntities list all region entity
	ListRegionEntities(ctx context.Context) ([]*models.RegionEntity, error)
	// GetRegionEntity get region entity, todo(gjq) add cache
	GetRegionEntity(ctx context.Context, regionName string) (*models.RegionEntity, error)
	GetRegionByID(ctx context.Context, id uint) (*models.RegionEntity, error)
	GetRegionByName(ctx context.Context, name string) (*models.Region, error)
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *models.Region) error
	// ListByRegionSelectors list region by tags
	ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (models.RegionParts, error)
	// DeleteByID delete region by id
	DeleteByID(ctx context.Context, id uint) error
}

type manager struct {
	regionDAO   regiondao.DAO
	registryDAO registrydao.DAO
	tagDAO      tagdao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		regionDAO:   regiondao.NewDAO(db),
		registryDAO: registrydao.NewDAO(db),
		tagDAO:      tagdao.NewDAO(db),
	}
}

func (m *manager) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	return m.regionDAO.Create(ctx, region)
}

func (m *manager) ListAll(ctx context.Context) ([]*models.Region, error) {
	return m.regionDAO.ListAll(ctx)
}

func (m *manager) ListRegionEntities(ctx context.Context) (ret []*models.RegionEntity, err error) {
	var regions []*models.Region
	regions, err = m.regionDAO.ListAll(ctx)
	if err != nil {
		return
	}

	for _, region := range regions {
		tags, err := m.tagDAO.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &models.RegionEntity{
			Region: region,
			Tags:   tags,
		})
	}
	return
}

func (m *manager) GetRegionEntity(ctx context.Context,
	regionName string) (*models.RegionEntity, error) {
	region, err := m.regionDAO.GetRegion(ctx, regionName)
	if err != nil {
		return nil, err
	}

	registry, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	return &models.RegionEntity{
		Region:   region,
		Registry: registry,
	}, nil
}

func (m *manager) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	_, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return err
	}
	// todo do more filed validation, for example ingressDomain must be format of the domain name
	return m.regionDAO.UpdateByID(ctx, id, region)
}

func (m *manager) getRegistryByRegion(ctx context.Context, region *models.Region) (*registrymodels.Registry, error) {
	registry, err := m.registryDAO.GetByID(ctx, region.RegistryID)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

func (m *manager) ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (
	models.RegionParts, error) {
	return m.regionDAO.ListByRegionSelectors(ctx, selectors)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.regionDAO.DeleteByID(ctx, id)
}

func (m *manager) GetRegionByID(ctx context.Context, id uint) (*models.RegionEntity, error) {
	region, err := m.regionDAO.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	registry, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	tags, err := m.tagDAO.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
	if err != nil {
		return nil, err
	}

	return &models.RegionEntity{
		Region:   region,
		Registry: registry,
		Tags:     tags,
	}, nil
}

func (m *manager) GetRegionByName(ctx context.Context, name string) (*models.Region, error) {
	return m.regionDAO.GetRegion(ctx, name)
}
