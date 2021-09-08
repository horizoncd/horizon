package dao

import (
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(db *gorm.DB, group *models.Group) (int, error)
	//Update(db *gorm.DB, group *models.Group) error
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(db *gorm.DB, group *models.Group) (int, error) {
	// todo 判断同名

	create := db.Create(group)

	return create.CreateBatchSize, nil
}
