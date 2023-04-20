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

	corecommon "github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/common"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/user/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	ListByEmail(ctx context.Context, emails []string) ([]*models.User, error)
	GetByIDs(ctx context.Context, userID []uint) ([]*models.User, error)
	List(ctx context.Context, query *q.Query) (int64, []*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	UpdateByID(ctx context.Context, id uint, newUser *models.User) (*models.User, error)
	GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d *dao) Create(ctx context.Context, user *models.User) (*models.User, error) {
	result := d.db.WithContext(ctx).Create(user)

	if result.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrInsertFailed(herrors.UserInDB, "failed to insert"),
			"failed to insert user(%#v): err = %v", user, result.Error.Error())
	}

	return user, result.Error
}

func (d *dao) List(ctx context.Context, query *q.Query) (int64, []*models.User, error) {
	var users []*models.User
	tx := d.db.Table("tb_user")
	if query != nil {
		for k, v := range query.Keywords {
			switch k {
			case corecommon.UserQueryName:
				tx = tx.Where("name like ?", fmt.Sprintf("%%%v%%", v))
			case corecommon.UserQueryType:
				tx = tx.Where("user_type in ?", v)
			}
		}
	}

	var total int64
	tx.Count(&total)

	if query != nil {
		tx = tx.Limit(query.Limit()).Offset(query.Offset())
	}
	res := tx.Scan(&users)

	err := res.Error
	if err != nil {
		return 0, nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "get user failed"),
			"get user failed:\n"+
				"query = %v\n err = %v", query, err)
	}

	if res.RowsAffected == 0 {
		return 0, make([]*models.User, 0), nil
	}
	return total, users, nil
}

func (d *dao) GetByIDs(ctx context.Context, userID []uint) ([]*models.User, error) {
	var users []*models.User
	result := d.db.WithContext(ctx).Raw(common.UserGetByID, userID).Scan(&users)
	if result.Error != nil {
		return nil,
			perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "failed to list users"),
				"failed to list users(%#v): err = %v", userID, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error) {
	var user models.User
	result := d.db.WithContext(ctx).Raw(common.UserQueryByOIDC, oidcType, email, models.UserTypeCommon).Scan(&user)
	if result.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "failed to get user"),
			"failed to get user(oidcType = %v, email = %v): err = %v",
			oidcType, email, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) ListByEmail(ctx context.Context, emails []string) ([]*models.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	var users []*models.User
	result := d.db.WithContext(ctx).Raw(common.UserListByEmail, emails, models.UserTypeCommon).Scan(&users)
	if result.Error != nil {
		return nil, perror.Wrapf(herrors.NewErrListFailed(herrors.UserInDB, "failed to list user"),
			"failed to list user(emails = %#v): err = %v", emails, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	res := d.db.Where("id = ?", id).First(&user)

	err := res.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perror.Wrapf(herrors.NewErrNotFound(herrors.UserInDB, "user not found"),
				"user not found:\n"+
					"id = %v\nerr = %v", id, err)
		}
		return nil, perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "failed to get user"),
			"failed to get user\n"+
				"id = %v\nerr = %v", id, err)
	}

	return &user, nil
}

func (d *dao) UpdateByID(ctx context.Context, id uint, newUser *models.User) (*models.User, error) {
	var user *models.User
	err := d.db.
		Transaction(
			func(tx *gorm.DB) error {
				res := tx.Where("id = ?", id).Select("admin", "banned").Updates(newUser)
				if res.Error != nil {
					return perror.Wrapf(herrors.NewErrUpdateFailed(herrors.UserInDB, "failed to update user"),
						"failed to update user\n"+
							"id = %v\nerr = %v", id, res.Error)
				}

				res = tx.Where("id = ?", id).First(&user)
				if err := res.Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return perror.Wrapf(herrors.NewErrNotFound(herrors.UserInDB, "user not found"),
							"user not found:\n"+
								"id = %v\nerr = %v", id, err)
					}
					return perror.Wrapf(herrors.NewErrGetFailed(herrors.UserInDB, "failed to get user"),
						"failed to get user\n"+
							"id = %v\nerr = %v", id, err)
				}
				return nil
			})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *dao) GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error) {
	var u *models.User
	result := d.db.Table("tb_user").
		Joins("join tb_idp_user l on l.user_id = tb_user.id").
		Joins("join tb_identity_provider i on i.id = l.idp_id").
		Where("tb_user.email = ?", email).
		Where("i.name = ?", idp).
		Select("tb_user.*").First(&u)
	if err := result.Error; err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.UserInDB, "failed to find user in db"),
			"failed to get user: err = %v", err)
	}
	return u, nil
}

func (d *dao) DeleteUser(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Exec(common.UserDeleteByID, id)
	return result.Error
}
