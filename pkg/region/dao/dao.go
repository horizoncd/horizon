package dao

import (
	"context"
	"fmt"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"

	"g.hz.netease.com/horizon/lib/orm"
	appregionmodels "g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/common"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/region/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create a region
	Create(ctx context.Context, region *models.Region) (*models.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*models.Region, error)
	// GetRegion get a region by name
	GetRegion(ctx context.Context, regionName string) (*models.Region, error)
	// GetRegionByID get a region by id
	GetRegionByID(ctx context.Context, id uint) (*models.Region, error)
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *models.Region) error
	// DeleteByID delete region by id
	DeleteByID(ctx context.Context, id uint) error
	// ListByRegionSelectors list region by tags
	ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (models.RegionParts, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct {
}

func (d *dao) ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (
	models.RegionParts, error) {
	if len(selectors) == 0 {
		return models.RegionParts{}, nil
	}
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var conditions []string
	var params []interface{}
	params = append(params, tagmodels.TypeRegion)
	for _, selector := range selectors {
		conditions = append(conditions, "(tg.tag_key = ? and tg.tag_value in ?)")
		params = append(params, selector.Key, selector.Values)
	}
	params = append(params, len(selectors))
	tagCondition := fmt.Sprintf("(%s)", strings.Join(conditions, " or "))
	var regionParts models.RegionParts
	result := db.Raw(fmt.Sprintf(common.RegionListByTags, tagCondition), params...).Scan(&regionParts)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regionParts, nil
}

// GetRegionByID implements DAO
func (*dao) GetRegionByID(ctx context.Context, id uint) (*models.Region, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var region models.Region
	result := db.Raw(common.RegionGetByID, id).First(&region)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.RegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// check en exist
	regionInDB, err := d.GetRegionByID(ctx, id)
	if err != nil {
		return err
	}

	// can only update displayName, server, Certificate, ingressDomain、harborID
	regionInDB.DisplayName = region.DisplayName
	regionInDB.Server = region.Server
	regionInDB.Certificate = region.Certificate
	regionInDB.IngressDomain = region.IngressDomain
	regionInDB.HarborID = region.HarborID
	regionInDB.Disabled = region.Disabled
	result := db.Save(regionInDB)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.RegionInDB, result.Error.Error())
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
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, result.Error
}

// DeleteByID implements DAO
func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	db, res := orm.FromContext(ctx)
	if res != nil {
		return res
	}

	// check region exist
	regionInDB, res := d.GetRegionByID(ctx, id)
	if res != nil {
		return res
	}

	// check if there are clusters using the region
	var count int64
	result := db.Raw(common.ClusterCountByRegionName, regionInDB.Name).Scan(&count)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
	}
	if count > 0 {
		return herrors.ErrRegionUsedByClusters
	}

	// remove related resources from different tables
	err := db.Transaction(func(tx *gorm.DB) error {
		// remove records from applicationRegion table
		result := tx.Where("region_name = ?", regionInDB.Name).Delete(&appregionmodels.ApplicationRegion{})
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		// remove records from environmentRegion table
		result = tx.Where("region_name = ?", regionInDB.Name).Delete(&envregionmodels.EnvironmentRegion{})
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		// remove records from tag table
		result = tx.Where("resource_id = ? and resource_type = ?", regionInDB.ID, tagmodels.TypeRegion).
			Delete(&tagmodels.Tag{})
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		// remove region itself
		result = tx.Delete(&models.Region{}, id)
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		return nil
	})

	return err
}
