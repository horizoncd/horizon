package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
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
		return nil, herrors.NewErrListFailed(herrors.ClusterTagInDB, result.Error.Error())
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
		result := db.Exec(common.ClusterTagDeleteAllByClusterID, clusterID)
		if result.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.ClusterTagInDB, result.Error.Error())
		}
		return nil
	}

	// 2. delete tags which keys not in the new keys
	tagKeys := make([]string, 0)
	for _, tag := range tags {
		tagKeys = append(tagKeys, tag.Key)
	}
	if err := db.Exec(common.ClusterTagDeleteByClusterIDAndKeys, clusterID, tagKeys).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.ClusterTagInDB, err.Error())
	}

	// 3. add new tags
	result := db.Clauses(clause.OnConflict{
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
		return herrors.NewErrInsertFailed(herrors.ClusterTagInDB, result.Error.Error())
	}
	return nil
}
