package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/registry/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create a harbor
	Create(ctx context.Context, harbor *models.Registry) (uint, error)
	// UpdateByID update a harbor
	UpdateByID(ctx context.Context, id uint, harbor *models.Registry) error
	// DeleteByID delete a harbor by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Registry, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) ([]*models.Registry, error)
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) Create(ctx context.Context, registry *models.Registry) (uint, error) {
	result := d.db.WithContext(ctx).Create(registry)

	if result.Error != nil {
		return 0, herrors.NewErrCreateFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return registry.ID, nil
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Registry, error) {
	var registry models.Registry
	result := d.db.WithContext(ctx).Where("id = ?", id).First(&registry)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.RegistryInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return &registry, nil
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Registry, error) {
	var registries []*models.Registry
	result := d.db.WithContext(ctx).Find(&registries)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return registries, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, registry *models.Registry) error {
	registry.ID = id
	result := d.db.WithContext(ctx).Save(registry)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	_, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// check if any region use the harbor
	var count int64
	result := d.db.WithContext(ctx).Model(&regionmodels.Region{}).
		Where("registry_id = ?", id).Where("deleted_ts = 0").Count(&count)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegistryInDB, result.Error.Error())
	}
	if count > 0 {
		return herrors.ErrHarborUsedByRegions
	}

	result = d.db.WithContext(ctx).Delete(&models.Registry{}, id)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.RegistryInDB, result.Error.Error())
	}

	return nil
}
