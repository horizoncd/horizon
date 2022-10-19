package dao

import (
	"context"
	"errors"
	"fmt"

	corecommon "g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

type DAO interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ListByEmail(ctx context.Context, emails []string) ([]*models.User, error)
	GetByIDs(ctx context.Context, userID []uint) ([]models.User, error)
	List(ctx context.Context, query *q.Query) (int64, []*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	UpdateByID(ctx context.Context, id uint, newUser *models.User) (*models.User, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

type dao struct{ db *gorm.DB }

func (d *dao) Create(ctx context.Context, user *models.User) (*models.User, error) {
	result := d.db.WithContext(ctx).Create(user)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.UserInDB, result.Error.Error())
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

func (d *dao) GetByIDs(ctx context.Context, userID []uint) ([]models.User, error) {
	var users []models.User
	result := d.db.WithContext(ctx).Raw(common.UserGetByID, userID).Scan(&users)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error) {
	var user models.User
	result := d.db.WithContext(ctx).Raw(common.UserQueryByOIDC, oidcType, email).Scan(&user)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.UserInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (d *dao) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := d.db.WithContext(ctx).Raw(common.UserQueryByEmail, email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perror.Wrapf(
				herrors.NewErrNotFound(herrors.UserInDB, "user not found"),
				"user not found:\n"+
					"email = %s\nerr = %v", email, err)
		}
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.UserInDB, "failed to get user"),
			"failed to get user:\n"+
				"email = %s\nerr = %v", email, err)
	}
	return &user, nil
}

func (d *dao) ListByEmail(ctx context.Context, emails []string) ([]*models.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	var users []*models.User
	result := d.db.WithContext(ctx).Raw(common.UserListByEmail, emails).Scan(&users)
	if result.Error != nil {
		return nil, herrors.NewErrListFailed(herrors.UserInDB, result.Error.Error())
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
