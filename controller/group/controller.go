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

var (
	Ctl = NewController()
)

const (
	ErrCodeNotFound     = errors.ErrorCode("GroupNotFound")
	ErrGroupHasChildren = errors.ErrorCode("GroupHasChildren")
	Type                = "group"
	ParentID            = "parent_id"
)

type Controller interface {
	CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*GChild, error)
	GetByPath(ctx context.Context, path string) (*GChild, error)
	Transfer(ctx context.Context, id, newParentID uint) error
	UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error)
	GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error)
	SearchGroups(ctx context.Context, id uint, filter string) ([]*GChild, int64, error)
	SearchChildren(ctx context.Context, id uint, filter string) ([]*GChild, int64, error)
}

type controller struct {
	groupManager group.Manager
}

func (c *controller) GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*GChild, int64, error) {
	return c.GetSubGroups(ctx, id, pageNumber, pageSize)
}

func (c *controller) SearchGroups(ctx context.Context, id uint, filter string) ([]*GChild, int64, error) {
	if filter == "" {
		return c.GetSubGroups(ctx, id, common.DefaultPageNumber, common.DefaultPageSize)
	}

	groupsByNames, err := c.groupManager.GetByNameFuzzily(ctx, filter)
	if err != nil {
		return nil, 0, err
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

	traversalIDsToGChildMap := formatTraversalIDsToGChildMap(groupsByIDs)

	// organize struct of search result
	parentIDToGChildMap := make(map[uint][]*GChild)
	firstLevelGChildren := make([]*GChild, 0)
	for i := range groupsByIDs {
		// reverse order because of the match logic
		g := groupsByIDs[len(groupsByIDs)-i-1]
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

func (c *controller) SearchChildren(ctx context.Context, id uint, filter string) ([]*GChild, int64, error) {
	return c.SearchGroups(ctx, id, filter)
}

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

func (c *controller) UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error {
	const op = "group *controller: update group basic info by id"

	groupEntity := convertUpdateGroupToGroup(updateGroup)
	groupEntity.ID = id

	rowsAffected, err := c.groupManager.UpdateBasic(ctx, groupEntity)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
	}

	return nil
}

func (c *controller) Transfer(ctx context.Context, id, newParentID uint) error {
	err := c.groupManager.Transfer(ctx, id, newParentID)
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error) {
	groupEntity := convertNewGroupToGroup(newGroup)

	return c.groupManager.Create(ctx, groupEntity)
}

func (c *controller) GetByPath(ctx context.Context, path string) (*GChild, error) {
	const op = "group *controller: get group by path"

	// path: /a/b => {a, b}
	paths := strings.Split(path[1:], "/")
	groups, err := c.groupManager.GetByPaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	traversalIDsToGChildMap := formatTraversalIDsToGChildMap(groups)

	for _, v := range traversalIDsToGChildMap {
		// path pointing to a group
		if v.FullPath == path {
			return v, nil
		}
	}

	return nil, errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the path: %s", path))
}

func (c *controller) GetByID(ctx context.Context, id uint) (*GChild, error) {
	const op = "group *controller: get group by id"

	groupEntity, err := c.groupManager.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
		}
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	groups, err := c.groupManager.GetByTraversalIDs(ctx, groupEntity.TraversalIDs)
	if err != nil {
		return nil, errors.E(op, fmt.Sprintf("failed to get the group matching the id: %d", id))
	}

	fullPath, fullName := formatFullPathAndFullName(groups)
	return convertGroupToGChild(groupEntity, fullName, fullPath, Type), nil
}

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

func NewController() Controller {
	return &controller{
		groupManager: group.Mgr,
	}
}
