package group

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"
)

var (
	// Mgr is the global group manager
	Mgr = New()

	ErrPathConflict = errors.New("path conflict")
	ErrHasChildren  = errors.New("children exist, cannot be deleted")
)

type Manager interface {
	Create(ctx context.Context, group *models.Group) (uint, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Group, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error)
	GetByIDsOrderByIDDesc(ctx context.Context, ids []uint) ([]*models.Group, error)
	GetByTraversalIDs(ctx context.Context, traversalIDs string) ([]*models.Group, error)
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	UpdateBasic(ctx context.Context, group *models.Group) error
	ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error)
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
	Transfer(ctx context.Context, id, newParentID uint) error
}

type manager struct {
	dao dao.DAO
}

func (m manager) Transfer(ctx context.Context, id, newParentID uint) error {
	group, err := m.GetByID(ctx, id)
	if err != nil {
		return err
	}
	pGroup, err := m.GetByID(ctx, newParentID)
	if err != nil {
		return err
	}

	return m.dao.Transfer(ctx, group.TraversalIDs, fmt.Sprintf("%s,%d", pGroup.TraversalIDs, group.ID))
}

func (m manager) GetByIDsOrderByIDDesc(ctx context.Context, ids []uint) ([]*models.Group, error) {
	return m.dao.GetByIDsOrderByIDDesc(ctx, ids)
}

func (m manager) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	return m.dao.GetByPaths(ctx, paths)
}

func (m manager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error) {
	return m.dao.GetByIDs(ctx, ids)
}

// GetByTraversalIDs traversalIDs: 1,2,3
func (m manager) GetByTraversalIDs(ctx context.Context, traversalIDs string) ([]*models.Group, error) {
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]uint, len(splitIds))
	for i, id := range splitIds {
		ii, _ := strconv.Atoi(id)
		ids[i] = uint(ii)
	}

	return m.GetByIDs(ctx, ids)
}

func (m manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	return m.dao.GetByNameFuzzily(ctx, name)
}

func (m manager) Create(ctx context.Context, group *models.Group) (uint, error) {
	var pGroup *models.Group
	var err error
	if group.ParentID == 0 {
		group.ParentID = -1
	} else {
		// check if parent exists
		pGroup, err = m.dao.GetByID(ctx, uint(group.ParentID))
		if err != nil {
			return 0, err
		}
	}

	// check if there's a record with the same parentId and name
	err = m.dao.CheckNameUnique(ctx, group)
	if err != nil {
		return 0, err
	}
	// check if there's a record with the same parentId and path
	err = m.dao.CheckPathUnique(ctx, group)
	if err != nil {
		return 0, err
	}

	id, err := m.dao.Create(ctx, group)
	if err != nil {
		return 0, err
	}

	// update traversal_ids, like 1; 1,2,3
	var traversalIDs string
	if group.ParentID == -1 {
		traversalIDs = strconv.Itoa(int(id))
	} else {
		traversalIDs = fmt.Sprintf("%s,%d", pGroup.TraversalIDs, id)
	}
	err = m.dao.UpdateTraversalIDs(ctx, id, traversalIDs)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m manager) Delete(ctx context.Context, id uint) error {
	record, err := m.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if record == nil {
		return gorm.ErrRecordNotFound
	}

	count, err := m.dao.CountByParentID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrHasChildren
	}
	// todo check application children exist

	return m.dao.Delete(ctx, id)
}

func (m manager) GetByID(ctx context.Context, id uint) (*models.Group, error) {
	return m.dao.GetByID(ctx, id)
}

func (m manager) UpdateBasic(ctx context.Context, group *models.Group) error {
	record, err := m.GetByID(ctx, group.ID)
	if err != nil {
		return err
	}
	if record == nil {
		return gorm.ErrRecordNotFound
	}

	// check if there's record with the same parentId and name
	err = m.dao.CheckNameUnique(ctx, group)
	if err != nil {
		return err
	}
	// check if there's a record with the same parentId and path
	err = m.dao.CheckPathUnique(ctx, group)
	if err != nil {
		return err
	}

	return m.dao.UpdateBasic(ctx, group)
}

func (m manager) ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error) {
	return m.dao.ListWithoutPage(ctx, query)
}

func (m manager) List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error) {
	return m.dao.List(ctx, query)
}

func New() Manager {
	return &manager{dao: dao.New()}
}
