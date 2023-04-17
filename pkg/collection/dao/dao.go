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
	resourceType string,
) (*models.Collection, error) {
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
	resourceType string,
) (*models.Collection, error) {
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
	ids []uint,
) ([]models.Collection, error) {
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
