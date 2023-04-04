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
	"fmt"
	"strings"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type UserLinkDAO interface {
	ListByUserID(ctx context.Context, uid uint) ([]*models.UserLink, error)
	GetByID(ctx context.Context, id uint) (*models.UserLink, error)
	DeleteByID(ctx context.Context, id uint) error
	CreateLink(ctx context.Context, link *models.UserLink) (*models.UserLink, error)
	GetByIDPAndSub(ctx context.Context, id uint, sub string) (*models.UserLink, error)
}

// NewUserLinkDAO returns an instance of the default UserLinkDAO
func NewUserLinkDAO(db *gorm.DB) UserLinkDAO {
	return &userLinkDAO{db: db}
}

type userLinkDAO struct{ db *gorm.DB }

func (d userLinkDAO) ListByUserID(ctx context.Context, uid uint) ([]*models.UserLink, error) {
	var links []*models.UserLink
	res := d.db.WithContext(ctx).
		Where("deleted_ts = 0").
		Where("user_id = ?", uid).
		Find(&links)
	err := res.Error
	if err != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(
			herrors.UserLinkInDB, "failed to get links"),
			"failed to get links:\n"+
				"user id = %d\nerr = %v", uid, err)
	}
	return links, nil
}

func (d userLinkDAO) GetByID(ctx context.Context, id uint) (*models.UserLink, error) {
	var link *models.UserLink
	err := d.db.Table("tb_idp_user").
		Where("id = ?", id).
		Where("deleted_ts = 0").
		First(&link).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.TemplateInDB, err.Error()),
				fmt.Sprintf("failed to find link: id = %d", id))
		}
		return nil, perror.Wrap(herrors.NewErrGetFailed(herrors.TemplateInDB, err.Error()),
			fmt.Sprintf("failed to get link: id = %d", id))
	}
	return link, err
}

func (d userLinkDAO) DeleteByID(ctx context.Context, id uint) error {
	var link *models.UserLink
	err := d.db.Table("tb_idp_user").
		Delete(&link, id).Error
	if err != nil {
		return perror.Wrap(herrors.NewErrDeleteFailed(herrors.TemplateInDB, err.Error()),
			fmt.Sprintf("failed to delete link: id = %d", id))
	}
	return nil
}

func (d userLinkDAO) GetByIDPAndSub(ctx context.Context, id uint, sub string) (*models.UserLink, error) {
	var link *models.UserLink
	err := d.db.WithContext(ctx).
		Table("tb_idp_user").
		Where("idp_id = ?", id).
		Where("sub = ?", sub).
		Where("deleted_ts = 0").
		First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.UserLinkInDB, "link not found"),
			"link not found:\n"+
				"idp id = %d\nerr = %v", id, err)
	}
	return link, nil
}

func (d userLinkDAO) CreateLink(ctx context.Context, link *models.UserLink) (*models.UserLink, error) {
	err := d.db.WithContext(ctx).Create(&link).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, perror.Wrapf(herrors.ErrDuplicatedKey,
				"failed to create link(%#v): err = %v", link, err.Error())
		}
		return nil, perror.Wrapf(herrors.NewErrCreateFailed(herrors.UserLinkInDB, "failed to create user link"),
			"failed to create user link:\n"+
				"err = %v", err)
	}
	return link, err
}
