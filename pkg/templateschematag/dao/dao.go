package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/templateschematag/models"
	"gorm.io/gorm/clause"
)

type DAO interface {
	// ListByClusterID List cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d dao) ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusterTags []*models.ClusterTemplateSchemaTag
	result := db.Raw(common.ClusterTemplateSchemaTagListByClusterID, clusterID).Scan(&clusterTags)

	if result.Error != nil {
		return nil, result.Error
	}

	return clusterTags, nil
}

func (d dao) UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// 1. if tags is empty, delete all tags
	if len(tags) == 0 {
		return db.Exec(common.ClusterTemplateSchemaTagDeleteAllByClusterID, clusterID).Error
	}

	// 2. delete tags which keys not in the new keys
	tagKeys := make([]string, 0)
	for _, tag := range tags {
		tagKeys = append(tagKeys, tag.Key)
	}
	if err := db.Exec(common.ClusterTemplateSchemaTagDeleteByClusterIDAndKeys, clusterID, tagKeys).Error; err != nil {
		return err
	}

	// 3. add new tags
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "cluster_id",
			}, {
				Name: "tag_key",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"tag_value"}),
	}).Create(tags).Error
}
