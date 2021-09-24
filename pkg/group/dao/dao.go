package dao

import (
	"context"
	"errors"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/models"
)

var (
	ErrPathConflict = errors.New("path conflict")
)

type DAO interface {
	CheckNameUnique(ctx context.Context, group *models.Group) error
	CheckPathUnique(ctx context.Context, group *models.Group) error
	Create(ctx context.Context, group *models.Group) (uint, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*models.Group, error)
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	CountByParentID(ctx context.Context, parentId uint) (int64, error)
	UpdateBasic(ctx context.Context, group *models.Group) error
	UpdateTraversalIDs(ctx context.Context, id uint, traversalIDs string) error
	ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error)
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) CountByParentID(ctx context.Context, parentId uint) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	result := db.Raw(common.GroupCountByParentID, parentId).Scan(&count)

	return count, result.Error
}

func (d *dao) UpdateTraversalIDs(ctx context.Context, id uint, traversalIDs string) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.GroupUpdateTraversalIDs, traversalIDs, id)

	return result.Error
}

func (d *dao) GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByIDsOrderByIDDesc, ids).Scan(&groups)

	return groups, result.Error
}

func (d *dao) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByPaths, paths).Scan(&groups)

	return groups, result.Error
}

func (d *dao) GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByIDs, ids).Scan(&groups)

	return groups, result.Error
}

// CheckPathUnique todo check application table too
func (d *dao) CheckPathUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	queryResult := models.Group{}
	result := db.Raw(common.GroupQueryByParentIDAndPathLimit1, group.ParentID, group.Path).Scan(&queryResult)

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

func (d *dao) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByNameFuzzily, name).Scan(groups)

	return groups, result.Error
}

// CheckNameUnique todo check application table too
func (d *dao) CheckNameUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	queryResult := models.Group{}
	result := db.Raw(common.GroupQueryByParentIDAndNameLimit1, group.ParentID, group.Name).Scan(&queryResult)

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

	result := db.Create(group)

	return group.ID, result.Error
}

// Delete can only delete a group that doesn't have any children
func (d *dao) Delete(ctx context.Context, id uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.GroupDelete, id)

	return result.Error
}

func (d *dao) Get(ctx context.Context, id uint) (*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var group models.Group
	result := db.Raw(common.GroupQueryByID, id).First(&group)

	return &group, result.Error
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

// UpdateBasic just update base info, not including transfer function
func (d *dao) UpdateBasic(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Raw(common.GroupUpdateBasic, group.Name, group.Path, group.Description, group.VisibilityLevel)

	return result.Error
}
