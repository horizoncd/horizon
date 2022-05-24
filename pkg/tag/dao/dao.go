package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"gorm.io/gorm/clause"
)

type DAO interface {
	// ListByResourceTypeID List tags by resourceType and resourceID
	ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error)
	// UpsertByResourceTypeID upsert tags
	UpsertByResourceTypeID(ctx context.Context, resourceType string, resourceID uint, tags []*models.Tag) error
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d dao) ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var tags []*models.Tag
	result := db.Raw(common.TagListByResourceTypeID, resourceType, resourceID).Scan(&tags)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.TagInDB, result.Error.Error())
	}

	return tags, nil
}

func (d dao) UpsertByResourceTypeID(ctx context.Context, resourceType string,
	resourceID uint, tags []*models.Tag) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// 1. if tags is empty, delete all tags
	if len(tags) == 0 {
		result := db.Exec(common.TagDeleteAllByResourceTypeID, resourceType, resourceID)
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
	if err := db.Exec(common.TagDeleteByResourceTypeIDAndKeys, resourceType, resourceID, tagKeys).Error; err != nil {
		return herrors.NewErrDeleteFailed(herrors.TagInDB, err.Error())
	}

	// 3. add new tags
	result := db.Clauses(clause.OnConflict{
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
