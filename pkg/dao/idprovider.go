// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"
	"errors"

	"github.com/horizoncd/horizon/core/common/idp"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type IDProviderDAO interface {
	List(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
	Create(ctx context.Context, idp *models.IdentityProvider) (*models.IdentityProvider, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error)
	Update(ctx context.Context, id uint, param *models.IdentityProvider) (*models.IdentityProvider, error)
	GetByCondition(ctx context.Context, condition q.Query) (*models.IdentityProvider, error)
}

type idProviderDAO struct {
	db *gorm.DB
}

var provider = models.IdentityProvider{}

func NewIDProviderDAO(db *gorm.DB) IDProviderDAO {
	return &idProviderDAO{
		db: db,
	}
}

func (d *idProviderDAO) List(ctx context.Context) ([]*models.IdentityProvider, error) {
	idps := make([]*models.IdentityProvider, 0)
	err := d.db.Find(&idps).Error
	if err != nil {
		return nil,
			perror.Wrap(herrors.NewErrGetFailed(herrors.IdentityProviderInDB, "failed to list err"),
				err.Error())
	}
	return idps, nil
}

func (d *idProviderDAO) GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error) {
	var res *models.IdentityProvider
	if err := d.db.First(&res, &models.IdentityProvider{Name: name}).Error; err != nil {
		return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.IdentityProviderInDB,
			err.Error()), "idp named %s not found", name)
	}
	return res, nil
}

func (d *idProviderDAO) Create(ctx context.Context,
	idp *models.IdentityProvider) (*models.IdentityProvider, error) {
	err := d.db.Create(idp).Error
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrCreateFailed(herrors.IdentityProviderInDB, err.Error()),
			"idp info = %v", idp)
	}
	return idp, nil
}

func (d *idProviderDAO) Delete(ctx context.Context, id uint) error {
	if err := d.db.Delete(&provider, id).Error; err != nil {
		return perror.Wrapf(
			herrors.NewErrDeleteFailed(herrors.IdentityProviderInDB, err.Error()),
			"idp id = %d", id)
	}
	return nil
}

func (d *idProviderDAO) GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error) {
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

func (d *idProviderDAO) GetByCondition(ctx context.Context,
	condition q.Query) (*models.IdentityProvider, error) {
	tx := d.db.Model(&provider)
	for k, v := range condition.Keywords {
		switch k {
		case idp.QueryName:
			tx.Where("name = ?", v)
		}
	}
	provider := models.IdentityProvider{}
	err := tx.First(&provider).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.IdentityProviderInDB, ""),
				"idp not found:\n"+
					"condition = %#v\n err = %v", condition, err)
		}
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.IdentityProviderInDB, ""),
			"idp get failed: \n"+
				"condition = %#v\n err = %v", condition, err)
	}
	return &provider, nil
}

func (d *idProviderDAO) Update(ctx context.Context,
	id uint, param *models.IdentityProvider) (*models.IdentityProvider, error) {
	res := &models.IdentityProvider{}
	err := d.db.Model(&provider).Where("id = ?", id).
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
