package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/harbor/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create a harbor
	Create(ctx context.Context, harbor *models.Harbor) (uint, error)
	// UpdateByID update a harbor
	UpdateByID(ctx context.Context, id uint, harbor *models.Harbor) error
	// DeleteByID delete a harbor by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Harbor, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) ([]*models.Harbor, error)
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) Create(ctx context.Context, harbor *models.Harbor) (uint, error) {
	result := d.db.WithContext(ctx).Create(harbor)

	if result.Error != nil {
		return 0, herrors.NewErrCreateFailed(herrors.HarborInDB, result.Error.Error())
	}

	return harbor.ID, nil
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Harbor, error) {
	var harbor models.Harbor
	result := d.db.WithContext(ctx).Raw(common.HarborGetByID, id).First(&harbor)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.HarborInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.HarborInDB, result.Error.Error())
	}

	return &harbor, nil
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Harbor, error) {
	var harbors []*models.Harbor
	result := d.db.WithContext(ctx).Raw(common.HarborListAll).Scan(&harbors)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.HarborInDB, result.Error.Error())
	}

	return harbors, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, harbor *models.Harbor) error {
	harborInDB, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	harborInDB.Name = harbor.Name
	harborInDB.Server = harbor.Server
	harborInDB.Token = harbor.Token
	harborInDB.PreheatPolicyID = harbor.PreheatPolicyID
	result := d.db.WithContext(ctx).Save(harborInDB)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.HarborInDB, result.Error.Error())
	}

	return nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	_, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// check if any region use the harbor
	var regions []*regionmodels.Region
	result := d.db.WithContext(ctx).Raw(common.RegionGetByHarborID, id).Scan(&regions)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.HarborInDB, result.Error.Error())
	}
	if len(regions) > 0 {
		return herrors.ErrHarborUsedByRegions
	}

	result = d.db.WithContext(ctx).Delete(&models.Harbor{}, id)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.HarborInDB, result.Error.Error())
	}

	return nil
}
