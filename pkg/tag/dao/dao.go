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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/tag/models"
)

type DAO interface {
	// ListByResourceTypeID Lists tags by resourceType and resourceID
	ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error)
	// ListByResourceTypeIDs ListByResourceTypeID Lists tags by resourceType and resourceIDs
	// if distinct enabled, tags only contains tag_key and tag_value
	ListByResourceTypeIDs(ctx context.Context, resourceType string, resourceIDs []uint,
		deduplicate bool) ([]*models.Tag, error)
	// UpsertByResourceTypeID upsert tags
	UpsertByResourceTypeID(ctx context.Context, resourceType string, resourceID uint, tags []*models.Tag) error
	CreateMetatags(ctx context.Context, metatags []*models.Metatag) error
	GetMetatagKeys(ctx context.Context) ([]string, error)
	GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error)
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d dao) ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := d.db.Transaction(func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Find(&tags)

		if result.Error != nil {
			return herrors.NewErrListFailed(herrors.TagInDB, result.Error.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (d dao) ListByResourceTypeIDs(ctx context.Context, resourceType string,
	resourceID []uint, deduplicate bool) ([]*models.Tag, error) {
	var tags []*models.Tag

	querySQL := common.TagListByResourceTypeIDs
	if deduplicate {
		querySQL = common.TagListDistinctByResourceTypeIDs
	}

	result := d.db.WithContext(ctx).Raw(querySQL, resourceType, resourceID).Scan(&tags)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.TagInDB, result.Error.Error())
	}

	return tags, nil
}

func (d dao) UpsertByResourceTypeID(ctx context.Context, resourceType string,
	resourceID uint, tags []*models.Tag) error {
	// 1. if tags is empty, delete all tags
	if len(tags) == 0 {
		result := d.db.WithContext(ctx).Exec(common.TagDeleteAllByResourceTypeID, resourceType, resourceID)
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.TagInDB, result.Error.Error())
		}
		return nil
	}

	// 2. delete tags which keys not in the new keys
	tagKeys := make([]string, 0)
	for _, tag := range tags {
		tagKeys = append(tagKeys, tag.Key)
	}
	if err := d.db.WithContext(ctx).Exec(common.TagDeleteByResourceTypeIDAndKeys, resourceType,
		resourceID, tagKeys).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.TagInDB, err.Error())
	}

	// 3. add new tags
	result := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "resource_type",
			}, {
				Name: "resource_id",
			}, {
				Name: "tag_key",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"tag_value"}),
	}).Create(tags)

	if result.Error != nil {
		return herrors.NewErrInsertFailed(herrors.TagInDB, result.Error.Error())
	}
	return nil
}

func (d dao) CreateMetatags(ctx context.Context, metatags []*models.Metatag) error {
	result := d.db.WithContext(ctx).Create(&metatags)
	if result.Error != nil {
		return herrors.NewErrInsertFailed(herrors.MetatagInDB, result.Error.Error())
	}

	return nil
}

func (d dao) GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error) {
	var metatags []*models.Metatag
	result := d.db.WithContext(ctx).Table("tb_metatag").Where("tag_key = ?", key).Find(&metatags)
	if result.Error != nil {
		return nil, result.Error
	}
	return metatags, nil
}

func (d dao) GetMetatagKeys(ctx context.Context) ([]string, error) {
	var keys []string
	result := d.db.WithContext(ctx).Table("tb_metatag").Distinct("tag_key").Find(&keys)
	if result.Error != nil {
		return nil, result.Error
	}
	return keys, nil
}
