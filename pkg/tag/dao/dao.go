package dao

import (
	"context"

	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/tag/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DAO interface {
	// ListByResourceTypeID Lists tags by resourceType and resourceID
	ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error)
	// ListByResourceTypeID Lists tags by resourceType and resourceIDs
	// if distinct enabled, tags only contains tag_key and tag_value
	ListByResourceTypeIDs(ctx context.Context, resourceType string, resourceIDs []uint,
		deduplicate bool) ([]*models.Tag, error)
	// UpsertByResourceTypeID upsert tags
	UpsertByResourceTypeID(ctx context.Context, resourceType string, resourceID uint, tags []*models.Tag) error
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d dao) ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error) {
	var tags []*models.Tag
	result := d.db.WithContext(ctx).Raw(common.TagListByResourceTypeID, resourceType, resourceID).Scan(&tags)

	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.TagInDB, result.Error.Error())
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
