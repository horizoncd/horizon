package manager

import (
	"context"
	"errors"

	"g.hz.netease.com/horizon/lib/q"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
)

var (
	// Mgr is the global group manager
	Mgr = New()

	// ErrHasChildren used when delete a group which still has some children
	ErrHasChildren = errors.New("children exist, cannot be deleted")
	// ErrConflictWithApplication conflict with the application
	ErrConflictWithApplication = errors.New("name or path is in conflict with application")
)

const (
	// _updateAt one of the field of the group table
	_updateAt = "updated_at"

	// _parentID one of the field of the group table
	_parentID = "parent_id"
)

type Manager interface {
	// Create a group
	Create(ctx context.Context, group *models.Group) (uint, error)
	// Delete a group by id
	Delete(ctx context.Context, id uint) (int64, error)
	// GetByID get a group by id
	GetByID(ctx context.Context, id uint) (*models.Group, error)
	// GetByIDs get groups by ids
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error)
	// GetByPaths get groups by paths
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	// GetByNameFuzzily get groups that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	// UpdateBasic update basic info of a group
	UpdateBasic(ctx context.Context, group *models.Group) error
	// GetSubGroupsUnderParentIDs get subgroups under the given parent groups without paging
	GetSubGroupsUnderParentIDs(ctx context.Context, parentIDs []uint) ([]*models.Group, error)
	// Transfer move a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// GetSubGroups get subgroups of a parent group, order by updateTime desc by default with paging
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*models.Group, int64, error)
	// GetByNameOrPathUnderParent get by name or path under a specified parent
	GetByNameOrPathUnderParent(ctx context.Context, name, path string, parentID uint) ([]*models.Group, error)
}

type manager struct {
	groupDAO       groupdao.DAO
	applicationDAO applicationdao.DAO
}

func (m manager) GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*models.Group, int64, error) {
	query := formatListGroupQuery(id, pageNumber, pageSize)
	return m.groupDAO.List(ctx, query)
}

func New() Manager {
	return &manager{
		groupDAO:       groupdao.NewDAO(),
		applicationDAO: applicationdao.NewDAO(),
	}
}

func (m manager) Transfer(ctx context.Context, id, newParentID uint) error {
	return m.groupDAO.Transfer(ctx, id, newParentID)
}

func (m manager) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	return m.groupDAO.GetByPaths(ctx, paths)
}

func (m manager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error) {
	return m.groupDAO.GetByIDs(ctx, ids)
}

func (m manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	return m.groupDAO.GetByNameFuzzily(ctx, name)
}

func (m manager) Create(ctx context.Context, group *models.Group) (uint, error) {
	if err := m.checkApplicationExists(ctx, group); err != nil {
		return 0, err
	}

	id, err := m.groupDAO.Create(ctx, group)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m manager) Delete(ctx context.Context, id uint) (int64, error) {
	count, err := m.groupDAO.CountByParentID(ctx, id)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrHasChildren
	}
	// todo check application children exist

	return m.groupDAO.Delete(ctx, id)
}

func (m manager) GetByID(ctx context.Context, id uint) (*models.Group, error) {
	return m.groupDAO.GetByID(ctx, id)
}

func (m manager) UpdateBasic(ctx context.Context, group *models.Group) error {
	if err := m.checkApplicationExists(ctx, group); err != nil {
		return err
	}

	// check record exist
	_, err := m.groupDAO.GetByID(ctx, group.ID)
	if err != nil {
		return err
	}

	// check if there's record with the same parentID and name
	err = m.groupDAO.CheckNameUnique(ctx, group)
	if err != nil {
		return err
	}
	// check if there's a record with the same parentID and path
	err = m.groupDAO.CheckPathUnique(ctx, group)
	if err != nil {
		return err
	}

	return m.groupDAO.UpdateBasic(ctx, group)
}

func (m manager) GetSubGroupsUnderParentIDs(ctx context.Context, parentIDs []uint) ([]*models.Group, error) {
	query := q.New(q.KeyWords{
		_parentID: parentIDs,
	})
	return m.groupDAO.ListWithoutPage(ctx, query)
}

// checkApplicationExists check application is already exists under the same parent
func (m manager) checkApplicationExists(ctx context.Context, group *models.Group) error {
	apps, err := m.applicationDAO.GetByNamesUnderGroup(ctx,
		group.ParentID, []string{group.Name, group.Path})
	if err != nil {
		return err
	}
	if len(apps) > 0 {
		return ErrConflictWithApplication
	}
	return nil
}

func (m manager) GetByNameOrPathUnderParent(ctx context.Context,
	name, path string, parentID uint) ([]*models.Group, error) {
	return m.groupDAO.GetByNameOrPathUnderParent(ctx, name, path, parentID)
}

// formatListGroupQuery query info for listing groups under a parent group, order by updated_at desc by default
func formatListGroupQuery(id uint, pageNumber, pageSize int) *q.Query {
	query := q.New(q.KeyWords{
		_parentID: id,
	})
	query.PageNumber = pageNumber
	query.PageSize = pageSize
	// sort by updated_at desc default，let newer items be in head
	s := q.NewSort(_updateAt, true)
	query.Sorts = []*q.Sort{s}

	return query
}
