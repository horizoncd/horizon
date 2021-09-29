package group

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group"
	"g.hz.netease.com/horizon/util/errors"
	"gorm.io/gorm"
)

const (
	// RootGroupID id of the root group, which is not actually exists in the group table
	RootGroupID = 0
)

var (
	// Ctl global instance of the group controller
	Ctl = NewController()
)

const (
	// ErrCodeNotFound a kind of error code, returned when there's no group matching the given id
	ErrCodeNotFound = errors.ErrorCode("GroupNotFound")
	// ErrGroupHasChildren a kind of error code, returned when deleting a group which still has some children
	ErrGroupHasChildren = errors.ErrorCode("GroupHasChildren")
	// Type used to indicate the 'Child' is a group
	Type = "group"
	// ParentID used in formatting query of the 'ListGroup'
	ParentID = "parent_id"
)

type Controller interface {
	// CreateGroup add a group
	CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error)
	// Delete remove a group by the id
	Delete(ctx context.Context, id uint) error
	// GetByID get a group by the id
	GetByID(ctx context.Context, id uint) (*GChild, error)
	// GetByFullPath get a group by the URLPath
	GetByFullPath(ctx context.Context, path string) (*GChild, error)
	// Transfer put a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// UpdateBasic update basic info of a group, including name, path, description and visibilityLevel
	UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error
	// GetSubGroups get subgroups of a group
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error)
	// GetChildren get children of a group, including subgroups and applications
	GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error)
	// SearchGroups search subGroups of a group
	SearchGroups(ctx context.Context, id uint, filter string) ([]*GChild, int64, error)
	// SearchChildren search children of a group, including subgroups and applications
	SearchChildren(ctx context.Context, id uint, filter string) ([]*GChild, int64, error)
}

type controller struct {
	groupManager group.Manager
}

// NewController initializes a new group controller
func NewController() Controller {
	return &controller{
		groupManager: group.Mgr,
	}
}

// GetChildren get children of a group, including subgroups and applications
func (c *controller) GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error) {
	return c.GetSubGroups(ctx, id, pageNumber, pageSize)
}

// SearchGroups search subGroups of a group
func (c *controller) SearchGroups(ctx context.Context, id uint, filter string) ([]*GChild, int64, error) {
	if filter == "" {
		return c.GetSubGroups(ctx, id, common.DefaultPageNumber, common.DefaultPageSize)
	}

	groupsByNames, err := c.groupManager.GetByNameFuzzily(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if groupsByNames == nil {
		return []*GChild{}, 0, nil
	}

	var ids []uint
	for _, g := range groupsByNames {
		split := strings.Split(g.TraversalIDs, ",")
		for _, s := range split {
			i, _ := strconv.Atoi(s)
			ids = append(ids, uint(i))
		}
	}

	groupsByIDs, err := c.groupManager.GetByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	traversalIDsToGChildMap := mappingTraversalIDsToGChild(groupsByIDs)

	// organize struct of search result
	parentIDToGChildMap := make(map[uint][]*GChild)
	firstLevelGChildren := make([]*GChild, 0)
	for i := len(groupsByIDs) - 1; i >= 0; i-- {
		// reverse order because of the match logic
		g := groupsByIDs[i]
		gChild := traversalIDsToGChildMap[g.TraversalIDs]

		// name match or children's names match
		if strings.Contains(g.Name, filter) || len(parentIDToGChildMap[g.ID]) > 0 {
			parentIDToGChildMap[g.ParentID] = append(parentIDToGChildMap[g.ParentID], gChild)
		}

		if v, ok := parentIDToGChildMap[gChild.ID]; ok {
			gChild.ChildrenCount = len(v)
			gChild.Children = v
		}

		// groups under the parent group
		if g.ParentID == id {
			firstLevelGChildren = append(firstLevelGChildren, gChild)
		}
	}

	return firstLevelGChildren, int64(len(firstLevelGChildren)), nil
}

// SearchChildren search children of a group, including subgroups and applications
func (c *controller) SearchChildren(ctx context.Context, id uint, filter string) ([]*GChild, int64, error) {
	return c.SearchGroups(ctx, id, filter)
}

// GetSubGroups get subgroups of a group
func (c *controller) GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error) {
	var pGChild *GChild
	if id > 0 {
		var err error
		pGChild, err = c.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}
	}

	// query subGroups
	query := formatListGroupQuery(id, pageNumber, pageSize)
	subGroups, count, err := c.groupManager.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// calculate childrenCount
	parentIDs := make([]uint, len(subGroups))
	for i, g := range subGroups {
		parentIDs[i] = g.ID
	}
	query = q.New(q.KeyWords{
		ParentID: parentIDs,
	})
	groups, err := c.groupManager.ListWithoutPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	childrenCountMap := map[uint]int{}
	for _, g := range groups {
		childrenCountMap[g.ParentID]++
	}

	// format GroupChild
	var gChildren = make([]*GChild, len(subGroups))
	for i, s := range subGroups {
		var fName, fPath string
		if id == 0 {
			fName = s.Name
			fPath = fmt.Sprintf("/%s", s.Path)
		} else {
			fName = fmt.Sprintf("%s / %s", pGChild.FullName, s.Name)
			fPath = fmt.Sprintf("%s/%s", pGChild.FullPath, s.Path)
		}
		gChild := convertGroupToGChild(s, fName, fPath, Type)
		gChild.ChildrenCount = childrenCountMap[gChild.ID]

		gChildren[i] = gChild
	}

	return gChildren, count, nil
}

// UpdateBasic update basic info of a group, including name, path, description and visibilityLevel
func (c *controller) UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error {
	groupEntity := convertUpdateGroupToGroup(updateGroup)
	groupEntity.ID = id

	err := c.groupManager.UpdateBasic(ctx, groupEntity)
	if err != nil {
		return err
	}

	return nil
}

// Transfer put a group under another parent group
func (c *controller) Transfer(ctx context.Context, id, newParentID uint) error {
	err := c.groupManager.Transfer(ctx, id, newParentID)
	if err != nil {
		return err
	}

	return nil
}

// CreateGroup add a group
func (c *controller) CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error) {
	groupEntity := convertNewGroupToGroup(newGroup)

	return c.groupManager.Create(ctx, groupEntity)
}

// GetByFullPath get a group by the URLPath
func (c *controller) GetByFullPath(ctx context.Context, path string) (*GChild, error) {
	const op = "group *controller: get group by path"

	// path: /a/b => {a, b}
	paths := strings.Split(path[1:], "/")
	groups, err := c.groupManager.GetByPaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	traversalIDsToGChildMap := mappingTraversalIDsToGChild(groups)

	for _, v := range traversalIDsToGChildMap {
		// path pointing to a group
		if v.FullPath == path {
			return v, nil
		}
	}

	return nil, errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the path: %s", path))
}

// GetByID get a group by the id
func (c *controller) GetByID(ctx context.Context, id uint) (*GChild, error) {
	const op = "group *controller: get group by id"

	groupEntity, err := c.groupManager.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
		}
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	groups, err := c.groupManager.GetByIDs(ctx, formatIDsFromTraversalIDs(groupEntity.TraversalIDs))
	if err != nil {
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	fullPath, fullName := formatFullPathAndFullName(groups)
	return convertGroupToGChild(groupEntity, fullName, fullPath, Type), nil
}

// Delete remove a group by the id
func (c *controller) Delete(ctx context.Context, id uint) error {
	const op = "group *controller: delete group by id"

	rowsAffected, err := c.groupManager.Delete(ctx, id)
	if err != nil {
		if err == group.ErrHasChildren {
			return errors.E(op, http.StatusBadRequest, ErrGroupHasChildren, group.ErrHasChildren.Error())
		}
		return errors.E(op, fmt.Sprintf("failed to delete the group matching the id: %d", id))
	}
	if rowsAffected == 0 {
		return errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
	}

	return nil
}
