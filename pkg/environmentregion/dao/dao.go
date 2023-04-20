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

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/environmentregion/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"gorm.io/gorm"
)

type DAO interface {
	// GetEnvironmentRegionByID ...
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	// ListAllEnvironmentRegions list all environmentRegions
	ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	// GetEnvironmentRegionByEnvAndRegion get
	GetEnvironmentRegionByEnvAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (*models.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by environment
	ListRegionsByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error)
	// ListEnabledRegionsByEnvironment list regions by environment that are enabled
	ListEnabledRegionsByEnvironment(ctx context.Context, env string) (regionmodels.RegionParts, error)
	// GetDefaultRegionByEnvironment get default regions by environment
	GetDefaultRegionByEnvironment(ctx context.Context, env string) (*models.EnvironmentRegion, error)
	// GetDefaultRegions get all default regions
	GetDefaultRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	// SetEnvironmentRegionToDefaultByID set region to default by id
	SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error
	// DeleteByID delete an environmentRegion by id
	DeleteByID(ctx context.Context, id uint) error
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) GetDefaultRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	var environmentRegion []*models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionsGetDefault).Scan(&environmentRegion)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environmentRegion, result.Error
}

func (d *dao) GetDefaultRegionByEnvironment(ctx context.Context, env string) (*models.EnvironmentRegion, error) {
	var environmentRegion models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionGetDefaultByEnv, env).First(&environmentRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return &environmentRegion, nil
}

func (d *dao) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	var environmentRegions []*models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionGetByEnvAndRegion, er.EnvironmentName,
		er.RegionName).Scan(&environmentRegions)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	if len(environmentRegions) > 0 {
		return nil, perror.Wrap(herrors.ErrPairConflict, "environmentRegion pair already exists")
	}

	result = d.db.WithContext(ctx).Create(er)
	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	return er, result.Error
}

func (d *dao) ListRegionsByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error) {
	var regions []*models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentListRegion, env).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return regions, result.Error
}

func (d *dao) ListEnabledRegionsByEnvironment(ctx context.Context, env string) (regionmodels.RegionParts, error) {
	var regions regionmodels.RegionParts
	result := d.db.WithContext(ctx).Raw(common.EnvironmentListEnabledRegion, env).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	if regions == nil {
		return regionmodels.RegionParts{}, nil
	}
	return regions, nil
}

func (d *dao) GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error) {
	var environmentRegion models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionGetByID, id).First(&environmentRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return &environmentRegion, nil
}

func (d *dao) GetEnvironmentRegionByEnvAndRegion(ctx context.Context,
	env, region string) (*models.EnvironmentRegion, error) {
	var environmentRegion models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionGet, env, region).First(&environmentRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	return &environmentRegion, nil
}

func (d *dao) SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error {
	region, err := d.GetEnvironmentRegionByID(ctx, id)
	if err != nil {
		return err
	}

	// get current default region
	currentDefaultRegion, err := d.GetDefaultRegionByEnvironment(ctx, region.EnvironmentName)
	if err != nil {
		// return if error is not HorizonErrNotFound type
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
	}

	if currentDefaultRegion == nil || currentDefaultRegion.ID != id {
		return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if currentDefaultRegion != nil {
				result := tx.Exec(common.EnvironmentRegionUnsetDefaultByID, currentDefaultRegion.ID)
				if result.Error != nil {
					return herrors.NewErrUpdateFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
				}
			}

			result := tx.Exec(common.EnvironmentRegionSetDefaultByID, id)
			if result.Error != nil {
				return herrors.NewErrUpdateFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
			}
			return nil
		})
	}

	return nil
}

func (d *dao) ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	var environmentRegions []*models.EnvironmentRegion
	result := d.db.WithContext(ctx).Raw(common.EnvironmentRegionListAll).Scan(&environmentRegions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environmentRegions, nil
}

// DeleteByID implements DAO
func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	_, err := d.GetEnvironmentRegionByID(ctx, id)
	if err != nil {
		return err
	}

	result := d.db.WithContext(ctx).Delete(&models.EnvironmentRegion{}, id)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return nil
}
