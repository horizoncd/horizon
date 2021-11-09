package service

import (
	"context"
	"g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/group/models"
)

const (
	// ChildTypeGroup used to indicate the 'Child' is a group
	ChildTypeGroup = "group"
	// ChildTypeApplication ...
	ChildTypeApplication = "application"
	// ChildTypeCluster ...
	ChildTypeCluster = "cluster"
	// RootGroupID id of the root group, which is not actually exists in the group table
	RootGroupID = 0
)

var (
	Svc = NewService()
)

type Service interface {
	// GetChildByID get a child by id
	GetChildByID(ctx context.Context, id uint) (*Child, error)
	// GetChildrenByIDs returns children map according to group ids
	GetChildrenByIDs(ctx context.Context, ids []uint) (map[uint]*Child, error)
}

type service struct {
	groupManager manager.Manager
}

func (s service) GetChildByID(ctx context.Context, id uint) (*Child, error) {
	group, err := s.groupManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	groups, err := s.groupManager.GetByIDs(ctx, manager.FormatIDsFromTraversalIDs(group.TraversalIDs))
	if err != nil {
		return nil, err
	}

	full := GenerateFullFromGroups(groups)

	return ConvertGroupToChild(group, full), nil
}

func (s service) GetChildrenByIDs(ctx context.Context, ids []uint) (map[uint]*Child, error) {
	var groupIDs []uint
	// childrenMap store result
	childrenMap := map[uint]*Child{}
	// groupMap store all queried groups
	groupMap := map[uint]*models.Group{}

	// 1.query groups
	groups, err := s.groupManager.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 2.query parent groups by traversal id, and store in map
	for _, group := range groups {
		for _, groupID := range manager.FormatIDsFromTraversalIDs(group.TraversalIDs) {
			groupMap[groupID] = nil
		}
	}
	for groupID := range groupMap {
		groupIDs = append(groupIDs, groupID)
	}
	parentGroups, err := s.groupManager.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	for i, group := range parentGroups {
		groupMap[group.ID] = parentGroups[i]
	}

	// 3.convert to children map
	for _, group := range groups {
		parentGroups = []*models.Group{}
		for _, id := range manager.FormatIDsFromTraversalIDs(group.TraversalIDs) {
			parentGroups = append(parentGroups, groupMap[id])
		}
		full := GenerateFullFromGroups(parentGroups)
		childrenMap[group.ID] = ConvertGroupToChild(group, full)
	}

	return childrenMap, nil
}

func NewService() Service {
	return &service{
		groupManager: manager.Mgr,
	}
}
