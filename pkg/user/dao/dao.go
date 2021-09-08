package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/user/models"
)

type DAO interface {
	Create(ctx context.Context, user *models.User) (int64, error)
	FindByOIDC(ctx context.Context, oidcID, oidcType string)(*models.User, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) FindByOIDC(ctx context.Context, oidcID, oidcType string)(*models.User, error){
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	result := db.Raw("SELECT * FROM user WHERE oidc_id = ? and oidc_type = ?", oidcID, oidcType).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (d *dao) Create(ctx context.Context, user *models.User) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(user)
	return result.RowsAffected, result.Error
}
