package dao

import (
	"context"
	"errors"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/idp/models"
	"gorm.io/gorm"
)

type DAO interface {
	List(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
	Create(ctx context.Context, idp *models.IdentityProvider) (*models.IdentityProvider, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error)
	Update(ctx context.Context, id uint, param *models.IdentityProvider) (*models.IdentityProvider, error)
}

type dao struct {
	db *gorm.DB
}

var model = models.IdentityProvider{}

func NewDAO(db *gorm.DB) DAO {
	return &dao{
		db: db,
	}
}

func (d *dao) List(ctx context.Context) ([]*models.IdentityProvider, error) {
	idps := make([]*models.IdentityProvider, 0)
	err := d.db.Find(&idps).Error
	if err != nil {
		return nil,
			perror.Wrap(herrors.NewErrGetFailed(herrors.IdentityProviderInDB, "failed to list err"),
				err.Error())
	}
	return idps, nil
}

func (d *dao) GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error) {
	var res *models.IdentityProvider
	if err := d.db.First(&res, &models.IdentityProvider{Name: name}).Error; err != nil {
		return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.IdentityProviderInDB,
			err.Error()), "idp named %s not found", name)
	}
	return res, nil
}

func (d *dao) Create(ctx context.Context,
	idp *models.IdentityProvider) (*models.IdentityProvider, error) {
	err := d.db.Create(idp).Error
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrCreateFailed(herrors.IdentityProviderInDB, err.Error()),
			"idp info = %v", idp)
	}
	return idp, nil
}

func (d *dao) Delete(ctx context.Context, id uint) error {
	if err := d.db.Delete(&model, id).Error; err != nil {
		return perror.Wrapf(
			herrors.NewErrDeleteFailed(herrors.IdentityProviderInDB, err.Error()),
			"idp id = %d", id)
	}
	return nil
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error) {
	res := &models.IdentityProvider{}
	err := d.db.Where("id = ?", id).First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perror.Wrapf(
				herrors.NewErrNotFound(herrors.IdentityProviderInDB, err.Error()),
				"idp with id = %d was not found", id)
		}
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.IdentityProviderInDB, err.Error()),
			"failed to get idp\n"+
				"id = %d\n"+
				"err = %v",
			id,
			err)
	}
	return res, nil
}

func (d *dao) Update(ctx context.Context,
	id uint, param *models.IdentityProvider) (*models.IdentityProvider, error) {
	res := &models.IdentityProvider{}
	err := d.db.Model(&model).Where("id = ?", id).
		Updates(param).Error
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrUpdateFailed(herrors.IdentityProviderInDB, err.Error()),
			"failed to update idp\n"+
				"idp id = %d\n"+
				"err = %v",
			id, err,
		)
	}
	return res, nil
}
