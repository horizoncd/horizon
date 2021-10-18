package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/region/models"
)

type DAO interface {
	// Create a region
	Create(ctx context.Context, region *models.Region) (*models.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*models.Region, error)
	// GetRegion get a region
	GetRegion(ctx context.Context, regionName string) (*models.Region, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct {
}

func (d *dao) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(region)

	return region, result.Error
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var regions []*models.Region
	result := db.Raw(common.RegionListAll).Scan(&regions)

	return regions, result.Error
}

func (d *dao) GetRegion(ctx context.Context, regionName string) (*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var region models.Region
	result := db.Raw(common.RegionGet, regionName).First(&region)

	return &region, result.Error
}
