package dao

import (
	"context"
	"github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/metatag/models"
	"gorm.io/gorm"
)

type DAO interface {
	CreateMetatags(ctx context.Context, metatags []*models.Metatag) error
	GetMetatagKeys(ctx context.Context) ([]string, error)
	GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d dao) CreateMetatags(ctx context.Context, metatags []*models.Metatag) error {
	result := d.db.WithContext(ctx).Create(&metatags)
	if result.Error != nil {
		return errors.NewErrInsertFailed(errors.DataMetatagInDB, result.Error.Error())
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
