package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/clustertag/models"
	"g.hz.netease.com/horizon/pkg/common"
	"gorm.io/gorm/clause"
)

type DAO interface {
	// ListByClusterID List cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTag) error
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d dao) ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTag, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var clusterTags []*models.ClusterTag
	result := db.Raw(common.ClusterTagListByClusterID, clusterID).Scan(&clusterTags)

	if result.Error != nil {
		return nil, result.Error
	}

	return clusterTags, nil
}

func (d dao) UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTag) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// 1. if tags is empty, delete all tags
	if len(tags) == 0 {
		return db.Exec(common.ClusterTagDeleteAllByClusterID, clusterID).Error
	}

	// 2. delete tags which keys not in the new keys
	tagKeys := make([]string, 0)
	for _, tag := range tags {
		tagKeys = append(tagKeys, tag.Key)
	}
	if err := db.Exec(common.ClusterTagDeleteByClusterIDAndKeys, clusterID, tagKeys).Error; err != nil {
		return err
	}

	// 3. add new tags
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "cluster_id",
			}, {
				Name: "key",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(tags).Error
}
