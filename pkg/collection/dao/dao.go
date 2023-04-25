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
	"errors"
	"fmt"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/collection/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, collection *models.Collection) (*models.Collection, error)
	DeleteByResource(ctx context.Context, userID uint, resourceID uint,
		resourceType string) (*models.Collection, error)
	GetByResource(ctx context.Context, userID uint, resourceID uint,
		resourceType string) (*models.Collection, error)
	List(ctx context.Context, userID uint, resourceType string, ids []uint) ([]models.Collection, error)
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d dao) Create(ctx context.Context, collection *models.Collection) (*models.Collection, error) {
	result := d.db.WithContext(ctx).Create(&collection)
	if result.Error != nil {
		return nil, herrors.NewErrCreateFailed(herrors.CollectionInDB,
			fmt.Sprintf("failed to create collection: %v", result.Error.Error()))
	}
	return collection, nil
}

func (d dao) GetByResource(ctx context.Context, userID uint, resourceID uint,
	resourceType string) (*models.Collection, error) {
	collection := models.Collection{}
	result := d.db.WithContext(ctx).Where("user_id = ?", userID).
		Where("resource_id = ?", resourceID).
		Where("resource_type = ?", resourceType).
		First(&collection)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, herrors.NewErrNotFound(herrors.CollectionInDB,
				fmt.Sprintf("collection not found: %v", result.Error.Error()))
		}
		return nil, herrors.NewErrGetFailed(herrors.CollectionInDB,
			fmt.Sprintf("could not get collection: %v", result.Error.Error()))
	}
	return &collection, nil
}

func (d dao) DeleteByResource(ctx context.Context, userID uint, resourceID uint,
	resourceType string) (*models.Collection, error) {
	collection := models.Collection{}
	result := d.db.WithContext(ctx).Where("user_id = ?", userID).
		Where("resource_id = ?", resourceID).
		Where("resource_type = ?", resourceType).Delete(&collection)
	if result.Error != nil {
		return nil,
			herrors.NewErrDeleteFailed(herrors.CollectionInDB,
				fmt.Sprintf("cloud not delete collection: %v", result.Error.Error()))
	}
	return &collection, nil
}

func (d dao) List(ctx context.Context, userID uint, resourceType string,
	ids []uint) ([]models.Collection, error) {
	var collections []models.Collection
	result := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("resource_type = ?", resourceType).
		Where("resource_id in ?", ids).Find(&collections)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.CollectionInDB,
			fmt.Sprintf("failed to list collections: %v", result.Error.Error()))
	}
	return collections, nil
}
