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

	"github.com/pkg/errors"
	"gorm.io/gorm"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/badge/models"
)

type DAO interface {
	List(ctx context.Context, resourceType string, resourceID uint) ([]*models.Badge, error)
	Create(ctx context.Context, badge *models.Badge) (*models.Badge, error)
	Update(ctx context.Context, badge *models.Badge) (*models.Badge, error)
	UpdateByName(ctx context.Context, resourceType string,
		resourceID uint, name string, badge *models.Badge) (*models.Badge, error)
	Get(ctx context.Context, id uint) (*models.Badge, error)
	GetByName(ctx context.Context, resourceType string, resourceID uint, name string) (*models.Badge, error)
	Delete(ctx context.Context, id uint) error
	DeleteByName(ctx context.Context, resourceType string, resourceID uint, name string) error

	DeleteByResource(ctx context.Context, resourceType string, resourceID uint) error
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) List(ctx context.Context, resourceType string, resourceID uint) ([]*models.Badge, error) {
	var badges []*models.Badge
	if err := d.db.WithContext(ctx).Where(
		"resource_type = ? AND resource_id = ?",
		resourceType, resourceID).Find(&badges).Error; err != nil {
		return nil, herrors.NewErrListFailed(herrors.BadgeInDB, err.Error())
	}
	return badges, nil
}

func (d *dao) Create(ctx context.Context, badge *models.Badge) (*models.Badge, error) {
	if err := d.db.WithContext(ctx).Create(badge).Error; err != nil {
		return nil, herrors.NewErrInsertFailed(herrors.BadgeInDB, err.Error())
	}
	return badge, nil
}

func (d *dao) Update(ctx context.Context, badge *models.Badge) (*models.Badge, error) {
	where := d.db.WithContext(ctx).Model(badge).Where("id = ?", badge.ID)
	if err := where.Updates(badge).Error; err != nil {
		return nil, herrors.NewErrUpdateFailed(herrors.BadgeInDB, err.Error())
	}
	if err := where.First(badge).Error; err != nil {
		return nil, herrors.NewErrGetFailed(herrors.BadgeInDB, err.Error())
	}
	return badge, nil
}

func (d *dao) UpdateByName(ctx context.Context, resourceType string,
	resourceID uint, name string, badge *models.Badge) (*models.Badge, error) {
	where := d.db.WithContext(ctx).Model(badge).
		Where("resource_type = ? AND resource_id = ? AND name = ?", resourceType, resourceID, name)
	if err := where.Updates(badge).Error; err != nil {
		return nil, herrors.NewErrUpdateFailed(herrors.BadgeInDB, err.Error())
	}
	if err := where.First(badge).Error; err != nil {
		return nil, herrors.NewErrGetFailed(herrors.BadgeInDB, err.Error())
	}
	return badge, nil
}

func (d *dao) Get(ctx context.Context, id uint) (*models.Badge, error) {
	var badge models.Badge
	if err := d.db.WithContext(ctx).First(&badge, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.BadgeInDB, err.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.BadgeInDB, err.Error())
	}
	return &badge, nil
}

func (d *dao) GetByName(ctx context.Context, resourceType string, resourceID uint, name string) (*models.Badge, error) {
	var badge models.Badge
	if err := d.db.WithContext(ctx).Where(
		"resource_type = ? AND resource_id = ? AND name = ?",
		resourceType, resourceID, name).First(&badge).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.BadgeInDB, err.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.BadgeInDB, err.Error())
	}
	return &badge, nil
}

func (d *dao) Delete(ctx context.Context, id uint) error {
	if err := d.db.WithContext(ctx).Delete(&models.Badge{}, id).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.BadgeInDB, err.Error())
	}
	return nil
}

func (d *dao) DeleteByName(ctx context.Context, resourceType string, resourceID uint, name string) error {
	if err := d.db.WithContext(ctx).Where(
		"resource_type = ? AND resource_id = ? AND name = ?",
		resourceType, resourceID, name).Delete(&models.Badge{}).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.BadgeInDB, err.Error())
	}
	return nil
}

func (d *dao) DeleteByResource(ctx context.Context, resourceType string, resourceID uint) error {
	if err := d.db.WithContext(ctx).Where(
		"resource_type = ? AND resource_id = ?",
		resourceType, resourceID).Delete(&models.Badge{}).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.BadgeInDB, err.Error())
	}
	return nil
}
