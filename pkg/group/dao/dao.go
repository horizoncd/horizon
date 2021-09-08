package dao

import (
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(db *gorm.DB, group *models.Group) (int64, error)
	Delete(db *gorm.DB, id int64) error
	Get(db *gorm.DB, id int64) (*models.Group, error)
	GetByPath(db *gorm.DB, path string) (*models.Group, error)
	Update(db *gorm.DB, group *models.Group) error
	GetChildren(db *gorm.DB, id int64) ([]*models.Group, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(db *gorm.DB, group *models.Group) (int64, error) {
	// todo 判断同名

	result := db.Create(group)

	return result.RowsAffected, result.Error
}

func (d *dao) Delete(db *gorm.DB, id int64) error {
	result := db.Where("id = ?", id).Delete(&models.Group{})

	return result.Error
}

func (d *dao) Get(db *gorm.DB, id int64) (*models.Group, error) {
	var group *models.Group
	result := db.First(&group, id)

	return group, result.Error
}

func (d *dao) GetByPath(db *gorm.DB, path string) (*models.Group, error) {
	var group *models.Group
	result := db.Find(&group, "path = ?", path)

	return group, result.Error
}

func (d *dao) GetChildren(db *gorm.DB, id int64) ([]*models.Group, error) {
	var groups []*models.Group
	result := db.Where("parentId = ?", id).Find(&groups)

	return groups, result.Error
}

func (d *dao) Update(db *gorm.DB, group *models.Group) error {
	// todo 判断同名

	result := db.Model(group).Select("Name", "Description", "VisibilityLevel").Updates(group)

	return result.Error
}
