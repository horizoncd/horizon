package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/idp/models"
	"gorm.io/gorm"
)

type DAO interface {
	ListIDP(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{
		db: db,
	}
}

func (d *dao) ListIDP(ctx context.Context) ([]*models.IdentityProvider, error) {
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
