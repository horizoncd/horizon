package dao

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"
)

var (
	ErrPathConflict = errors.New("path conflict")
	ErrHasChildren  = errors.New("children exist, cannot be deleted")
)

type DAO interface {
	CheckNameUnique(ctx context.Context, group *models.Group) error
	CheckPathUnique(ctx context.Context, group *models.Group) error
	Create(ctx context.Context, group *models.Group) (uint, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*models.Group, error)
	GetByPath(ctx context.Context, path string) (*models.Group, error)
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	GetByFullNamesRegexpFuzzily(ctx context.Context, names *[]string) ([]*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error)
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw("select * from `group` where id in ? order by id desc", ids).Scan(&groups)

	return groups, result.Error
}

func (d *dao) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw("select * from `group` where path in ?", paths).Scan(&groups)

	return groups, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group

	result := db.Find(&groups, ids)

	return groups, result.Error
}

// CheckPathUnique todo check application table too
func (d *dao) CheckPathUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	query := map[string]interface{}{
		"parent_id": group.ParentID,
		"path":      group.Path,
	}

	queryResult := &models.Group{}
	result := db.Model(&models.Group{}).Where(query).Find(queryResult)

	// update group conflict, has another record with the same parentId & path
	if group.ID > 0 && queryResult.ID > 0 && queryResult.ID != group.ID {
		return ErrPathConflict
	}

	// create group conflict
	if group.ID == 0 && result.RowsAffected > 0 {
		return ErrPathConflict
	}

	return nil
}

func (d *dao) GetByFullNamesRegexpFuzzily(ctx context.Context, names *[]string) ([]*models.Group, error) {
	if names == nil || (len(*names)) == 0 {
		return []*models.Group{}, nil
	}

	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	formatNames := make([]string, len(*names))
	for i, n := range *names {
		formatNames[i] = fmt.Sprintf("full_name like '%s%%'", n)
	}

	query := strings.Join(formatNames, " or ")
	// select * from `group` where full_name like '1%' or full_name like '2%' order by id desc
	qSQL := fmt.Sprintf("select * from `group` where %s order by id desc", query)
	result := db.Raw(qSQL).Scan(&groups)

	return groups, result.Error
}

func (d *dao) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Where("name LIKE ?", "%"+name+"%").Find(&groups)

	return groups, result.Error
}

// CheckNameUnique todo check application table too
func (d *dao) CheckNameUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	query := map[string]interface{}{
		"parent_id": group.ParentID,
		"name":      group.Name,
	}

	queryResult := &models.Group{}
	result := db.Model(&models.Group{}).Where(query).Find(queryResult)

	// update group conflict, has another record with the same parentId & name
	if group.ID > 0 && queryResult.ID > 0 && queryResult.ID != group.ID {
		return common.ErrNameConflict
	}

	// create group conflict
	if group.ID == 0 && result.RowsAffected > 0 {
		return common.ErrNameConflict
	}

	return nil
}

func (d *dao) Create(ctx context.Context, group *models.Group) (uint, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	// if group's parentId is 0, set to -1
	var pGroup *models.Group
	if group.ParentID == 0 {
		group.ParentID = -1
	} else {
		// check if parent exists
		pGroup, err = d.Get(ctx, uint(group.ParentID))
		if err != nil {
			return 0, err
		}
	}
	// check if there's a record with the same parentId and name
	err = d.CheckNameUnique(ctx, group)
	if err != nil {
		return 0, err
	}
	// check if there's a record with the same parentId and path
	err = d.CheckPathUnique(ctx, group)
	if err != nil {
		return 0, err
	}

	result := db.Create(group)
	if result.Error != nil {
		return 0, result.Error
	}
	// update traversal_ids, like 1; 1,2,3
	var traversalIDs string
	if group.ParentID == -1 {
		traversalIDs = strconv.Itoa(int(group.ID))
	} else {
		traversalIDs = fmt.Sprintf("%s,%d", pGroup.TraversalIDs, group.ID)
	}
	result = db.Model(&group).Update("traversal_ids", traversalIDs)

	return group.ID, result.Error
}

// Delete can only delete a group that doesn't have any children
func (d *dao) Delete(ctx context.Context, id uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	var groupChild *models.Group
	result := db.Where("parent_id = ?", id).First(&groupChild)
	if result.RowsAffected > 0 {
		return ErrHasChildren
	}
	// todo check application children exist

	result = db.Where("id = ?", id).Delete(&models.Group{})
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

func (d *dao) ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group

	sort := orm.FormatSortExp(query)
	result := db.Order(sort).Where(query.Keywords).Find(&groups)

	return groups, result.Error
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
	result := db.Order(sort).Where(query.Keywords).Offset(offset).Limit(query.PageSize).Find(&groups).
		Offset(-1).Count(&count)
	return groups, count, result.Error
}

// Update just update base info, not including transfer function
func (d *dao) Update(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// check if there's record with the same parentId and name
	err = d.CheckNameUnique(ctx, group)
	if err != nil {
		return err
	}
	// check if there's a record with the same parentId and path
	err = d.CheckPathUnique(ctx, group)
	if err != nil {
		return err
	}

	result := db.Model(group).Select("Name", "Description", "VisibilityLevel").Where("deleted_at is null").Updates(group)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return result.Error
}
