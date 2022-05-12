package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"

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
	// ListByNames list by names
	ListByNames(ctx context.Context, regionNames []string) ([]*models.Region, error)
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *models.Region) error
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct {
}

func (d *dao) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// can only update displayName, server, Certificate, ingressDomain、harborID
	// check en exist
	var regionInDB models.Region
	res := db.Find(&regionInDB, id)
	if res.RowsAffected == 0 {
		return herrors.NewErrNotFound(herrors.RegionInDB, "rows affected = 0")
	}
	if res.Error != nil {
		return herrors.NewErrGetFailed(herrors.RegionInDB, res.Error.Error())
	}

	regionInDB.DisplayName = region.DisplayName
	regionInDB.Server = region.Server
	regionInDB.Certificate = region.Certificate
	regionInDB.IngressDomain = region.IngressDomain
	regionInDB.HarborID = region.HarborID
	res = db.Save(regionInDB)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentInDB, res.Error.Error())
	}

	return nil
}

func (d *dao) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(region)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.RegionInDB, result.Error.Error())
	}

	return region, result.Error
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var regions []*models.Region
	result := db.Raw(common.RegionListAll).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regions, result.Error
}

func (d *dao) GetRegion(ctx context.Context, regionName string) (*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var region models.Region
	result := db.Raw(common.RegionGetByName, regionName).First(&region)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return &region, result.Error
}

func (d *dao) ListByNames(ctx context.Context, regionNames []string) ([]*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var regions []*models.Region
	result := db.Raw(common.RegionListByNames, regionNames).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return regions, result.Error
}
