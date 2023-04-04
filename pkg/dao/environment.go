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
	"sort"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	appregionmodels "github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type EnvironmentDAO interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *appregionmodels.Environment) (*appregionmodels.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*appregionmodels.Environment, error)
	// UpdateByID update environment by id
	UpdateByID(ctx context.Context, id uint, environment *appregionmodels.Environment) error
	// DeleteByID delete environment by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get environment by id
	GetByID(ctx context.Context, id uint) (*appregionmodels.Environment, error)
	// GetByName get environment by name
	GetByName(ctx context.Context, name string) (*appregionmodels.Environment, error)
}

type environmentDAO struct{ db *gorm.DB }

// NewEnvironmentDAO returns an instance of the default EnvironmentDAO
func NewEnvironmentDAO(db *gorm.DB) EnvironmentDAO {
	return &environmentDAO{db: db}
}

func (d *environmentDAO) UpdateByID(ctx context.Context, id uint, environment *appregionmodels.Environment) error {
	environmentInDB, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// set displayName and autoFree
	environmentInDB.DisplayName = environment.DisplayName
	res := d.db.WithContext(ctx).Save(&environmentInDB)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentInDB, res.Error.Error())
	}

	return nil
}

func (d *environmentDAO) CreateEnvironment(ctx context.Context,
	environment *appregionmodels.Environment) (*appregionmodels.Environment, error) {
	result := d.db.WithContext(ctx).Create(environment)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environment, result.Error
}

func (d *environmentDAO) ListAllEnvironment(ctx context.Context) ([]*appregionmodels.Environment, error) {
	var environments []*appregionmodels.Environment

	result := d.db.WithContext(ctx).Raw(common.EnvironmentListAll).Scan(&environments)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	sort.Sort(appregionmodels.EnvironmentList(environments))
	return environments, nil
}

func (d *environmentDAO) DeleteByID(ctx context.Context, id uint) error {
	environment, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// remove related resources from different tables
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// remove records from applicationRegion table
		res := tx.Where("environment_name = ?", environment.Name).Delete(&appregionmodels.ApplicationRegion{})
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, res.Error.Error())
		}

		// remove records from environmentRegion table
		res = tx.Where("environment_name = ?", environment.Name).Delete(&appregionmodels.EnvironmentRegion{})
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, res.Error.Error())
		}

		// remove environment itself
		res = tx.Delete(&appregionmodels.Environment{}, id)
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.EnvironmentInDB, res.Error.Error())
		}
		return nil
	})

	return err
}

func (d *environmentDAO) GetByID(ctx context.Context, id uint) (*appregionmodels.Environment, error) {
	var environment appregionmodels.Environment
	result := d.db.WithContext(ctx).Raw(common.EnvironmentGetByID, id).First(&environment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentInDB, result.Error.Error())
	}

	return &environment, nil
}

func (d *environmentDAO) GetByName(ctx context.Context, name string) (*appregionmodels.Environment, error) {
	var environment appregionmodels.Environment
	result := d.db.WithContext(ctx).Raw(common.EnvironmentGetByName, name).First(&environment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentInDB, result.Error.Error())
	}

	return &environment, nil
}
