// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"
	"fmt"
	"strings"

	hcommon "github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	appregionmodels "github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type RegionDAO interface {
	// Create a region
	Create(ctx context.Context, region *appregionmodels.Region) (*appregionmodels.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*appregionmodels.Region, error)
	// GetRegion get a region by name
	GetRegion(ctx context.Context, regionName string) (*appregionmodels.Region, error)
	// GetRegionByID get a region by id
	GetRegionByID(ctx context.Context, id uint) (*appregionmodels.Region, error)
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *appregionmodels.Region) error
	// DeleteByID delete region by id
	DeleteByID(ctx context.Context, id uint) error
	// ListByRegionSelectors list region by tags
	ListByRegionSelectors(ctx context.Context,
		selectors appregionmodels.RegionSelectors) (appregionmodels.RegionParts, error)
}

// NewRegionDAO returns an instance of the default RegionDAO
func NewRegionDAO(db *gorm.DB) RegionDAO {
	return &regionDAO{db: db}
}

type regionDAO struct {
	db *gorm.DB
}

func (d *regionDAO) ListByRegionSelectors(ctx context.Context, selectors appregionmodels.RegionSelectors) (
	appregionmodels.RegionParts, error) {
	if len(selectors) == 0 {
		return appregionmodels.RegionParts{}, nil
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
	var regionParts appregionmodels.RegionParts
	result := d.db.WithContext(ctx).Raw(fmt.Sprintf(common.RegionListByTags, tagCondition), params...).Scan(&regionParts)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regionParts, nil
}

// GetRegionByID implements RegionDAO
func (d *regionDAO) GetRegionByID(ctx context.Context, id uint) (*appregionmodels.Region, error) {
	var region appregionmodels.Region
	result := d.db.WithContext(ctx).Raw(common.RegionGetByID, id).First(&region)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.RegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, nil
}

func (d *regionDAO) UpdateByID(ctx context.Context, id uint, region *appregionmodels.Region) error {
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

func (d *regionDAO) Create(ctx context.Context, region *appregionmodels.Region) (*appregionmodels.Region, error) {
	result := d.db.WithContext(ctx).Create(region)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.RegionInDB, result.Error.Error())
	}

	return region, result.Error
}

func (d *regionDAO) ListAll(ctx context.Context) ([]*appregionmodels.Region, error) {
	var regions []*appregionmodels.Region
	result := d.db.WithContext(ctx).Raw(common.RegionListAll).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return regions, result.Error
}

func (d *regionDAO) GetRegion(ctx context.Context, regionName string) (*appregionmodels.Region, error) {
	var region appregionmodels.Region
	result := d.db.WithContext(ctx).Raw(common.RegionGetByName, regionName).First(&region)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, result.Error.Error())
	}

	return &region, nil
}

// DeleteByID implements RegionDAO
func (d *regionDAO) DeleteByID(ctx context.Context, id uint) error {
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
		result = tx.Where("region_name = ?", regionInDB.Name).Delete(&appregionmodels.EnvironmentRegion{})
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		// remove records from tag table
		result = tx.Where("resource_id = ? and resource_type = ?", regionInDB.ID, hcommon.ResourceRegion).
			Delete(&appregionmodels.Tag{})
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		// remove region itself
		result = tx.Delete(&appregionmodels.Region{}, id)
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, result.Error.Error())
		}

		return nil
	})

	return err
}
