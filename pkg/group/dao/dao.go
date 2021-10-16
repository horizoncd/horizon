package dao

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"gorm.io/gorm"
)

var (
	ErrPathConflict = errors.New("path conflict")
	ErrNameConflict = errors.New("name conflict")
)

type DAO interface {
	// CheckNameUnique check whether the name is unique
	CheckNameUnique(ctx context.Context, group *models.Group) error
	// CheckPathUnique check whether the path is unique
	CheckPathUnique(ctx context.Context, group *models.Group) error
	// Create a group
	Create(ctx context.Context, group *models.Group) (*models.Group, error)
	// Delete a group by id
	Delete(ctx context.Context, id uint) (int64, error)
	// GetByID get a group by id
	GetByID(ctx context.Context, id uint) (*models.Group, error)
	// GetByNameFuzzily get groups that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	// GetByIDNameFuzzily get groups that fuzzily matching the given name and id
	GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*models.Group, error)
	// GetByIDs get groups by ids
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error)
	// GetByPaths get groups by paths
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	// CountByParentID get the count of the records matching the given parentID
	CountByParentID(ctx context.Context, parentID uint) (int64, error)
	// UpdateBasic update basic info of a group
	UpdateBasic(ctx context.Context, group *models.Group) error
	// ListWithoutPage query groups without paging
	ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error)
	// List query groups with paging
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
	// ListChildren children of a group
	ListChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) ([]*models.GroupOrApplication, int64, error)
	// Transfer move a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint, userID uint) error
	// GetByNameOrPathUnderParent get by name or path under a specified parent
	GetByNameOrPathUnderParent(ctx context.Context, name, path string, parentID uint) ([]*models.Group, error)
}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByIDNameFuzzily, fmt.Sprintf("%%%d%%", id),
		fmt.Sprintf("%%%s%%", name)).Scan(&groups)

	return groups, result.Error
}

func (d *dao) ListChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) (
	[]*models.GroupOrApplication, int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var gas []*models.GroupOrApplication
	var count int64

	result := db.Raw(common.GroupQueryGroupChildren, parentID, parentID, pageSize, (pageNumber-1)*pageSize).Scan(&gas)
	if result.Error != nil {
		return nil, 0, err
	}

	result = db.Raw(common.GroupQueryGroupChildrenCount, parentID, parentID).Scan(&count)

	return gas, count, result.Error
}

func (d *dao) Transfer(ctx context.Context, id, newParentID uint, userID uint) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// check records exist
	group, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}
	pGroup, err := d.GetByID(ctx, newParentID)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// change parentID
		if err := tx.Exec(common.GroupUpdateParentID, newParentID, userID, id).Error; err != nil {
			return err
		}

		// update traversalIDs
		oldTIDs := group.TraversalIDs
		newTIDs := fmt.Sprintf("%s,%d", pGroup.TraversalIDs, group.ID)
		if err := tx.Exec(common.GroupUpdateTraversalIDsPrefix, oldTIDs, newTIDs, userID, oldTIDs+"%").Error; err != nil {
			return err
		}

		// commit when return nil
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *dao) CountByParentID(ctx context.Context, parentID uint) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	result := db.Raw(common.GroupCountByParentID, parentID).Scan(&count)

	return count, result.Error
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

func (d *dao) GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByIDs, ids).Scan(&groups)

	return groups, result.Error
}

func (d *dao) CheckPathUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	queryResult := models.Group{}
	result := db.Raw(common.GroupQueryByParentIDAndPath, group.ParentID, group.Path).First(&queryResult)

	// update group conflict, has another record with the same parentID & path
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
	result := db.Raw(common.GroupQueryByNameFuzzily, fmt.Sprintf("%%%s%%", name)).Scan(&groups)

	return groups, result.Error
}

func (d *dao) CheckNameUnique(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	queryResult := models.Group{}
	result := db.Raw(common.GroupQueryByParentIDAndName, group.ParentID, group.Name).First(&queryResult)

	// update group conflict, has another record with the same parentID & name
	if group.ID > 0 && queryResult.ID > 0 && queryResult.ID != group.ID {
		return ErrNameConflict
	}

	// create group conflict
	if group.ID == 0 && result.RowsAffected > 0 {
		return ErrNameConflict
	}

	return nil
}

func (d *dao) Create(ctx context.Context, group *models.Group) (*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var pGroup *models.Group
	// check if parent exists
	if group.ParentID > 0 {
		pGroup, err = d.GetByID(ctx, group.ParentID)
		if err != nil {
			return nil, err
		}
	}

	// check if there's a record with the same parentID and name
	err = d.CheckNameUnique(ctx, group)
	if err != nil {
		return nil, err
	}
	// check if there's a record with the same parentID and path
	err = d.CheckPathUnique(ctx, group)
	if err != nil {
		return nil, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// create, get id returned by the database
		if err := tx.Create(group).Error; err != nil {
			// rollback when error
			return err
		}

		// update traversalIDs
		id := group.ID
		var traversalIDs string
		if pGroup == nil {
			traversalIDs = strconv.Itoa(int(id))
		} else {
			traversalIDs = fmt.Sprintf("%s,%d", pGroup.TraversalIDs, id)
		}

		if err := tx.Exec(common.GroupUpdateTraversalIDs, traversalIDs, id).Error; err != nil {
			// rollback when error
			return err
		}

		// insert a record to member table
		member := &membermodels.Member{
			ResourceType: membermodels.TypeGroup,
			ResourceID:   id,
			//TODO(tom): where to place the role
			Role:         membermodels.Owner,
			MemberType:   membermodels.MemberUser,
			MemberNameID: group.CreatedBy,
			GrantBy:      group.UpdatedBy,
		}
		result := tx.Create(member)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("create member error")
		}

		// commit when return nil
		return nil
	})

	if err != nil {
		return nil, err
	}

	return group, nil
}

// Delete can only delete a group that doesn't have any children
func (d *dao) Delete(ctx context.Context, id uint) (int64, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Exec(common.GroupDelete, id)

	return result.RowsAffected, result.Error
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Group, error) {
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

// UpdateBasic just update base info, not contains transfer logic
func (d *dao) UpdateBasic(ctx context.Context, group *models.Group) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	result := db.Exec(common.GroupUpdateBasic, group.Name, group.Path, group.Description,
		group.VisibilityLevel, group.UpdatedBy, group.ID)

	return result.Error
}

func (d *dao) GetByNameOrPathUnderParent(ctx context.Context,
	name, path string, parentID uint) ([]*models.Group, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var groups []*models.Group
	result := db.Raw(common.GroupQueryByNameOrPathUnderParent, parentID, name, path).Scan(&groups)

	return groups, result.Error
}
