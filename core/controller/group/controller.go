package group

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/util/errors"

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

	// ChildTypeGroup used to indicate the 'Child' is a group
	ChildTypeGroup = "group"
	// ChildTypeApplication ...
	ChildTypeApplication = "application"
)

type Controller interface {
	// CreateGroup add a group
	CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error)
	// Delete remove a group by the id
	Delete(ctx context.Context, id uint) error
	// GetByID get a group by the id
	GetByID(ctx context.Context, id uint) (*Child, error)
	// GetByFullPath get a group by the URLPath
	GetByFullPath(ctx context.Context, path string) (*Child, error)
	// Transfer put a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// UpdateBasic update basic info of a group, including name, path, description and visibilityLevel
	UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error
	// GetSubGroups get subgroups of a group
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*Child, int64, error)
	// GetChildren get children of a group, including subgroups and applications
	GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*Child, int64, error)
	// SearchGroups search subGroups of a group
	SearchGroups(ctx context.Context, params *SearchParams) ([]*Child, int64, error)
	// SearchChildren search children of a group, including subgroups and applications
	SearchChildren(ctx context.Context, params *SearchParams) ([]*Child, int64, error)
}

type controller struct {
	groupManager       manager.Manager
	applicationManager appmanager.Manager
}

// NewController initializes a new group controller
func NewController() Controller {
	return &controller{
		groupManager:       manager.Mgr,
		applicationManager: appmanager.Mgr,
	}
}

// GetChildren get children of a group, including subgroups and applications
func (c *controller) GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*Child, int64, error) {
	return c.GetSubGroups(ctx, id, pageNumber, pageSize)
}

// SearchGroups search subGroups of a group
func (c *controller) SearchGroups(ctx context.Context, params *SearchParams) ([]*Child, int64, error) {
	if params.Filter == "" {
		return c.GetSubGroups(ctx, params.GroupID, params.PageNumber, params.PageSize)
	}

	// query groups by the name fuzzily
	groupsByNames, err := c.groupManager.GetByNameFuzzily(ctx, params.Filter)
	if err != nil {
		return nil, 0, err
	}
	if groupsByNames == nil {
		return []*Child{}, 0, nil
	}

	// query groups in ids (split groupsByNames's traversalIDs by ',')
	groups, err := c.formatGroupsInTraversalIDs(ctx, groupsByNames)
	if err != nil {
		return nil, 0, err
	}

	// generate children with level struct
	childrenWithLevelStruct := generateChildrenWithLevelStruct(params.GroupID, groups)

	// sort children by updatedAt desc
	sort.SliceStable(childrenWithLevelStruct, func(i, j int) bool {
		return childrenWithLevelStruct[i].UpdatedAt.After(childrenWithLevelStruct[j].UpdatedAt)
	})
	return childrenWithLevelStruct, int64(len(childrenWithLevelStruct)), nil
}

// SearchChildren search children of a group, including subgroups and applications
func (c *controller) SearchChildren(ctx context.Context, params *SearchParams) ([]*Child, int64, error) {
	return c.SearchGroups(ctx, params)
}

// GetSubGroups get subgroups of a group
func (c *controller) GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*Child, int64, error) {
	var pGChild *Child
	if id > 0 {
		var err error
		pGChild, err = c.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}
	}

	// query subGroups
	subGroups, count, err := c.groupManager.GetSubGroups(ctx, id, pageNumber, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// calculate childrenCount
	parentIDs := make([]uint, len(subGroups))
	for i, g := range subGroups {
		parentIDs[i] = g.ID
	}
	groups, err := c.groupManager.GetSubGroupsUnderParentIDs(ctx, parentIDs)
	if err != nil {
		return nil, 0, err
	}
	childrenCountMap := map[uint]int{}
	for _, g := range groups {
		childrenCountMap[g.ParentID]++
	}

	// format GroupChild
	var gChildren = make([]*Child, len(subGroups))
	for i, s := range subGroups {
		var fName, fPath string
		if id == 0 {
			fName = s.Name
			fPath = fmt.Sprintf("/%s", s.Path)
		} else {
			fName = fmt.Sprintf("%s / %s", pGChild.FullName, s.Name)
			fPath = fmt.Sprintf("%s/%s", pGChild.FullPath, s.Path)
		}
		child := convertGroupToChild(s, &Full{
			FullName: fName,
			FullPath: fPath,
		})
		child.ChildrenCount = childrenCountMap[child.ID]

		gChildren[i] = child
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
func (c *controller) GetByFullPath(ctx context.Context, path string) (*Child, error) {
	const op = "group *controller: get group by path"

	var errNotMatch = errors.E(op, http.StatusNotFound, ErrCodeNotFound,
		fmt.Sprintf("no resource matching the path: %s", path))

	// path: /a/b => {a, b}
	paths := strings.Split(path[1:], "/")
	groups, err := c.groupManager.GetByPaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	// get mapping between id and group
	idToGroup := generateIDToGroup(groups)

	// get mapping between id and full
	idToFull := generateIDToFull(groups)

	// 1. match group
	for k, v := range idToFull {
		// path pointing to a group
		if v.FullPath == path {
			g := idToGroup[k]
			child := convertGroupToChild(g, v)
			return child, nil
		}
	}

	// 2. match application
	if len(paths) < 2 {
		return nil, errNotMatch
	}
	app, err := c.applicationManager.GetByName(ctx, paths[len(paths)-1])
	if err != nil {
		return nil, err
	}
	if app != nil {
		appParentFull, ok := idToFull[app.GroupID]
		if ok && fmt.Sprintf("%v/%v", appParentFull.FullPath, app.Name) == path {
			return convertApplicationToChild(app, &Full{
				FullName: fmt.Sprintf("%v / %v", appParentFull.FullName, app.Name),
				FullPath: fmt.Sprintf("%v/%v", appParentFull.FullPath, app.Name),
			})
		}
	}

	// 3. todo match cluster
	if len(paths) < 3 {
		return nil, errNotMatch
	}

	return nil, errNotMatch
}

// GetByID get a group by the id
func (c *controller) GetByID(ctx context.Context, id uint) (*Child, error) {
	const op = "group *controller: get group by id"

	groupEntity, err := c.groupManager.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
		}
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	groups, err := c.groupManager.GetByIDs(ctx, manager.FormatIDsFromTraversalIDs(groupEntity.TraversalIDs))
	if err != nil {
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	full := generateFullFromGroups(groups)
	return convertGroupToChild(groupEntity, full), nil
}

// Delete remove a group by the id
func (c *controller) Delete(ctx context.Context, id uint) error {
	const op = "group *controller: delete group by id"

	rowsAffected, err := c.groupManager.Delete(ctx, id)
	if err != nil {
		if err == manager.ErrHasChildren {
			return errors.E(op, http.StatusBadRequest, ErrGroupHasChildren, manager.ErrHasChildren)
		}
		return errors.E(op, fmt.Sprintf("failed to delete the group matching the id: %d", id), err)
	}
	if rowsAffected == 0 {
		return errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
	}

	return nil
}

// formatGroupsInTraversalIDs query groups by ids (split traversalIDs by ',')
func (c *controller) formatGroupsInTraversalIDs(ctx context.Context, groups []*models.Group) ([]*models.Group, error) {
	var ids []uint
	for _, g := range groups {
		ids = append(ids, manager.FormatIDsFromTraversalIDs(g.TraversalIDs)...)
	}

	groupsByIDs, err := c.groupManager.GetByIDs(ctx, ids)
	if err != nil {
		return []*models.Group{}, err
	}

	return groupsByIDs, nil
}

// generateChildrenWithLevelStruct generate subgroups with level struct
func generateChildrenWithLevelStruct(groupID uint, groups []*models.Group) []*Child {
	// get mapping between id and full
	idToFull := generateIDToFull(groups)

	// first level children under the group
	firstLevelChildren := make([]*Child, 0)

	// record the mapping between parentID and children
	parentIDToChildren := make(map[uint][]*Child)

	// reverse the order
	sort.Sort(sort.Reverse(models.Groups(groups)))
	for _, g := range groups {
		// get fullName and fullPath by id
		full := idToFull[g.ID]
		child := convertGroupToChild(g, full)

		// record children of the group whose id is g.parentID
		parentIDToChildren[g.ParentID] = append(parentIDToChildren[g.ParentID], child)

		if v, ok := parentIDToChildren[g.ID]; ok {
			child.ChildrenCount = len(v)
			child.Children = v
		}

		if g.ParentID == groupID {
			firstLevelChildren = append(firstLevelChildren, child)
		}
	}

	return firstLevelChildren
}
