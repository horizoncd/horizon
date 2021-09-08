package dao

import (
	"g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(db *gorm.DB, user *models.User) (int64, error)
	FindByOIDC(db *gorm.DB, oidcID, oidcType string)(*models.User, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) FindByOIDC(db *gorm.DB, oidcID, oidcType string)(*models.User, error){
	var user models.User
	result := db.Raw("SELECT * FROM user WHERE oidc_id = ? and oidc_type = ?", oidcID, oidcType).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (d *dao) Create(db *gorm.DB, user *models.User) (int64, error) {
	result := db.Create(user)
	return result.RowsAffected, result.Error
}
