package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, group *models.Group) (uint, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*models.Group, error)
	GetByPath(ctx context.Context, path string) (*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, group *models.Group) (uint, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(group)

	return group.ID, result.Error
}

func (d *dao) Delete(ctx context.Context, id uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Where("id = ?", id).Delete(&models.Group{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return result.Error
}

func (d *dao) Get(ctx context.Context, id uint) (*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var group *models.Group
	result := db.First(&group, id)

	return group, result.Error
}

func (d *dao) GetByPath(ctx context.Context, path string) (*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var group *models.Group
	result := db.First(&group, "path = ?", path)

	return group, result.Error
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var groups []*models.Group

	sort := orm.FormatSortExp(query)
	offset := (query.PageNumber - 1) * query.PageSize
	var count int64
	result := db.Order(sort).Where(query.Keywords).Limit(query.PageSize).Offset(offset).Find(&groups).Count(&count)

	return groups, count, result.Error
}

func (d *dao) Update(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Model(group).Select("Name", "Description", "VisibilityLevel").Where("deleted_at is null").Updates(group)

	// todo modify children's path & fullName

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
