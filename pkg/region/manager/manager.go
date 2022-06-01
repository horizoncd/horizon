package manager

import (
	"context"
	"fmt"

	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
	"g.hz.netease.com/horizon/pkg/region/models"
	tagdao "g.hz.netease.com/horizon/pkg/tag/dao"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
)

var (
	// Mgr is the global region manager
	Mgr = New()
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
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *models.Region) error
	// ListByRegionSelectors list region by tags
	ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (models.RegionParts, error)
	// DeleteByID delete region by id
	DeleteByID(ctx context.Context, id uint) error
}

type manager struct {
	regionDAO regiondao.DAO
	harborDAO harbordao.DAO
	tagDAO    tagdao.DAO
}

func New() Manager {
	return &manager{
		regionDAO: regiondao.NewDAO(),
		harborDAO: harbordao.NewDAO(),
		tagDAO:    tagdao.NewDAO(),
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

	harborMap, err := m.getHarborMap(ctx)
	if err != nil {
		return nil, err
	}

	for _, region := range regions {
		harbor, ok := harborMap[region.HarborID]
		if !ok {
			return nil, fmt.Errorf("harbor with ID: %v of region: %v is not found",
				region.HarborID, region.Name)
		}
		tags, err := m.tagDAO.ListByResourceTypeID(ctx, tagmodels.TypeRegion, region.ID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &models.RegionEntity{
			Region: region,
			Harbor: harbor,
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

	harbor, err := m.getHarborByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	return &models.RegionEntity{
		Region: region,
		Harbor: harbor,
	}, nil
}

func (m *manager) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	_, err := m.getHarborByRegion(ctx, region)
	if err != nil {
		return err
	}
	// todo do more filed validation, for example ingressDomain must be format of the domain name
	return m.regionDAO.UpdateByID(ctx, id, region)
}

func (m *manager) getHarborMap(ctx context.Context) (map[uint]*harbormodels.Harbor, error) {
	harbors, err := m.harborDAO.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	harborMap := make(map[uint]*harbormodels.Harbor)
	for _, harbor := range harbors {
		harborMap[harbor.ID] = harbor
	}
	return harborMap, nil
}

func (m *manager) getHarborByRegion(ctx context.Context, region *models.Region) (*harbormodels.Harbor, error) {
	harbor, err := m.harborDAO.GetByID(ctx, region.HarborID)
	if err != nil {
		return nil, err
	}
	return harbor, nil
}

func (m *manager) ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (
	models.RegionParts, error) {
	return m.regionDAO.ListByRegionSelectors(ctx, selectors)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.regionDAO.DeleteByID(ctx, id)
}
