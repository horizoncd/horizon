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
	regionmodels "github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type RegistryDAO interface {
	// Create a registry
	Create(ctx context.Context, registry *regionmodels.Registry) (uint, error)
	// UpdateByID update a registry
	UpdateByID(ctx context.Context, id uint, registry *regionmodels.Registry) error
	// DeleteByID delete a registry by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*regionmodels.Registry, error)
	// ListAll list all registries
	ListAll(ctx context.Context) ([]*regionmodels.Registry, error)
}

type registryDAO struct{ db *gorm.DB }

// NewRegistryDAO returns an instance of the default RegistryDAO
func NewRegistryDAO(db *gorm.DB) RegistryDAO {
	return &registryDAO{db: db}
}

func (d *registryDAO) Create(ctx context.Context, registry *regionmodels.Registry) (uint, error) {
	result := d.db.WithContext(ctx).Create(registry)

	if result.Error != nil {
		return 0, herrors.NewErrCreateFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return registry.ID, nil
}

func (d *registryDAO) GetByID(ctx context.Context, id uint) (*regionmodels.Registry, error) {
	var registry regionmodels.Registry
	result := d.db.WithContext(ctx).Where("id = ?", id).First(&registry)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.RegistryInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return &registry, nil
}

func (d *registryDAO) ListAll(ctx context.Context) ([]*regionmodels.Registry, error) {
	var registries []*regionmodels.Registry
	result := d.db.WithContext(ctx).Find(&registries)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return registries, nil
}

func (d *registryDAO) UpdateByID(ctx context.Context, id uint, registry *regionmodels.Registry) error {
	result := d.db.WithContext(ctx).Where("id = ?", id).Select("Name", "Server", "Path",
		"Token", "InsecureSkipTLSVerify", "Kind").Updates(registry)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return nil
}

func (d *registryDAO) DeleteByID(ctx context.Context, id uint) error {
	_, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// check if any region use the registry
	var count int64
	result := d.db.WithContext(ctx).Model(&regionmodels.Region{}).
		Where("registry_id = ?", id).Where("deleted_ts = 0").Count(&count)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegistryInDB, result.Error.Error())
	}
	if count > 0 {
		return herrors.ErrRegistryUsedByRegions
	}

	result = d.db.WithContext(ctx).Delete(&regionmodels.Registry{}, id)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return nil
}
