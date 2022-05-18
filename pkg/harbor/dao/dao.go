package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"gorm.io/gorm"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/harbor/models"
)

type DAO interface {
	// Create a harbor
	Create(ctx context.Context, harbor *models.Harbor) (*models.Harbor, error)
	// Update update a harbor
	Update(ctx context.Context, harbor *models.Harbor) error
	// DeleteByID delete a harbor by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Harbor, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) ([]*models.Harbor, error)
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) Create(ctx context.Context, harbor *models.Harbor) (*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(harbor)

	if result.Error != nil {
		return nil, herrors.NewErrCreateFailed(herrors.HarborInDB, result.Error.Error())
	}

	return harbor, nil
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var harbor models.Harbor
	result := db.Raw(common.HarborGetByID, id).First(&harbor)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.HarborInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.HarborInDB, result.Error.Error())
	}

	return &harbor, nil
}

func (d *dao) ListAll(ctx context.Context) ([]*models.Harbor, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var harbors []*models.Harbor
	result := db.Raw(common.HarborListAll).Scan(&harbors)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.HarborInDB, result.Error.Error())
	}

	return harbors, nil
}

func (d *dao) Update(ctx context.Context, harbor *models.Harbor) error {
	_, err := d.GetByID(ctx, harbor.ID)
	if err != nil {
		return err
	}

	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Save(harbor)
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

	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// check if any region use the harbor
	var regions []*regionmodels.Region
	result := db.Raw(common.RegionGetByHarborID, id).Scan(&regions)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.HarborInDB, result.Error.Error())
	}
	if len(regions) > 0 {
		return herrors.ErrHarborCannotDelete
	}

	result = db.Delete(&models.Harbor{}, id)
	if result.Error != nil {
		return herrors.NewErrDeleteFailed(herrors.HarborInDB, result.Error.Error())
	}

	return nil
}
