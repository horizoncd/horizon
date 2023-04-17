package dao

import (
	"context"
	"fmt"
	"strings"

	herrors "github.com/horizoncd/horizon/core/errors"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"

	hcommon "github.com/horizoncd/horizon/core/common"
	appregionmodels "github.com/horizoncd/horizon/pkg/applicationregion/models"
	"github.com/horizoncd/horizon/pkg/common"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/region/models"
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

// NewDAO returns an instance of the default DAO.
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct {
	db *gorm.DB
}

func (d *dao) ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (
	models.RegionParts, error,
) {
	if len(selectors) == 0 {
		return models.RegionParts{}, nil
	}

	var conditions []string
	var params []interface{}
	params = append(params, hcommon.ResourceRegion)
	for _, selector := range selectors {
		conditions = append(conditions, "(tg.tag_key = ? and tg.tag_value in ?)")
		params = append(params, selector.Key, selector.Values)
	}
	params = append(params, len(selectors))
	tagCondition := fmt.Sprintf("(%s)", strings.Join(conditions, " or "))
	var regionParts models.RegionParts
	result := d.db.WithContext(ctx).Raw(fmt.Sprintf(common.RegionListByTags, tagCondition), params...).Scan(&regionParts)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regionParts, nil
}

// GetRegionByID implements DAO.
func (d *dao) GetRegionByID(ctx context.Context, id uint) (*models.Region, error) {
	var region models.Region
	result := d.db.WithContext(ctx).Raw(common.RegionGetByID, id).First(&region)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.RegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	// check en exist
	regionInDB, err := d.GetRegionByID(ctx, id)
	if err != nil {
		return err
	}

	// can only update displayName, server, Certificate, ingressDomain、prometheusURL, registryID
	regionInDB.DisplayName = region.DisplayName
	regionInDB.Server = region.Server
	regionInDB.Certificate = region.Certificate
	regionInDB.IngressDomain = region.IngressDomain
	regionInDB.PrometheusURL = region.PrometheusURL
	regionInDB.RegistryID = region.RegistryID
	regionInDB.Disabled = region.Disabled
	result := d.db.WithContext(ctx).Save(regionInDB)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.RegionInDB, result.Error.Error())
	}

	return nil
}

func (d *dao) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	result := d.db.WithContext(ctx).Create(region)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.RegionInDB, result.Error.Error())
	}

	return region, result.Error
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Region, error) {
	var regions []*models.Region
	result := d.db.WithContext(ctx).Raw(common.RegionListAll).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regions, result.Error
}

func (d *dao) GetRegion(ctx context.Context, regionName string) (*models.Region, error) {
	var region models.Region
	result := d.db.WithContext(ctx).Raw(common.RegionGetByName, regionName).First(&region)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, nil
}

// DeleteByID implements DAO.
func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	// check region exist
	regionInDB, res := d.GetRegionByID(ctx, id)
	if res != nil {
		return res
	}

	// check if there are clusters using the region
	var count int64
	result := d.db.WithContext(ctx).Raw(common.ClusterCountByRegionName, regionInDB.Name).Scan(&count)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
	}
	if count > 0 {
		return herrors.ErrRegionUsedByClusters
	}

	// remove related resources from different tables
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		result = tx.Where("resource_id = ? and resource_type = ?", regionInDB.ID, hcommon.ResourceRegion).
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
