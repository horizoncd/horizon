package service

import (
	"context"

	"g.hz.netease.com/horizon/pkg/group/manager"
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

func NewService() Service {
	return &service{
		groupManager: manager.Mgr,
	}
}
