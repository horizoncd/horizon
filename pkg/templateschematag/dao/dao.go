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
	"github.com/horizoncd/horizon/pkg/templateschematag/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DAO interface {
	// ListByClusterID Lists cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d dao) ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error) {
	var tags []*models.ClusterTemplateSchemaTag
	result := d.db.WithContext(ctx).Raw(common.ClusterTemplateSchemaTagListByClusterID, clusterID).Scan(&tags)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.TemplateSchemaTagInDB, result.Error.Error())
	}

	return tags, nil
}

func (d dao) UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error {
	// 1. if tags is empty, delete all tags
	if len(tags) == 0 {
		result := d.db.WithContext(ctx).Exec(common.ClusterTemplateSchemaTagDeleteAllByClusterID, clusterID)

		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.TemplateSchemaTagInDB, result.Error.Error())
		}
		return nil
	}

	// 2. delete tags which keys not in the new keys
	tagKeys := make([]string, 0)
	for _, tag := range tags {
		tagKeys = append(tagKeys, tag.Key)
	}
	if err := d.db.WithContext(ctx).Exec(common.ClusterTemplateSchemaTagDeleteByClusterIDAndKeys,
		clusterID, tagKeys).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.TemplateSchemaTagInDB, err.Error())
	}

	// 3. add new tags
	result := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "cluster_id",
			}, {
				Name: "tag_key",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"tag_value"}),
	}).Create(tags)

	if result.Error != nil {
		return herrors.NewErrCreateFailed(herrors.TemplateSchemaTagInDB, result.Error.Error())
	}
	return nil
}
