package dao

import (
	"context"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/user/models"
)

type DAO interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	// GetByOIDCMeta get user by oidcID and oidcType
	GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*models.User, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, user *models.User) (*models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(user)
	return user, result.Error
}

func (d *dao) GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*models.User, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	result := db.Raw(common.UserQueryByOIDC, oidcID, oidcType).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}
