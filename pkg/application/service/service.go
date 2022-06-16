package service

import (
	"context"
	"fmt"

	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
)

type Service interface {
	// GetByID get application with full name and full path by id
	GetByID(ctx context.Context, id uint) (*ApplicationDetail, error)
	// GetByIDs get application map with full name and full path by ids
	GetByIDs(ctx context.Context, ids []uint) (map[uint]*ApplicationDetail, error)
}

type service struct {
	groupService       groupservice.Service
	applicationManager applicationmanager.Manager
}

func (s service) GetByID(ctx context.Context, id uint) (*ApplicationDetail, error) {
	application, err := s.applicationManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	group, err := s.groupService.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, application.Name)
	fullName := fmt.Sprintf("%v/%v", group.FullName, application.Name)

	applicationDetail := &ApplicationDetail{
		*application,
		fullPath,
		fullName,
	}
	return applicationDetail, nil
}

func (s service) GetByIDs(ctx context.Context, ids []uint) (map[uint]*ApplicationDetail, error) {
	applicationMap := map[uint]*ApplicationDetail{}
	// 1. get applications
	applications, err := s.applicationManager.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 2. get groups for full path, full name
	var groupIDs []uint
	for _, application := range applications {
		groupIDs = append(groupIDs, application.GroupID)
	}
	groupMap, err := s.groupService.GetChildrenByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	// 3. add full path and full name
	for i, application := range applications {
		fullPath := fmt.Sprintf("%v/%v", groupMap[application.GroupID].FullPath, application.Name)
		fullName := fmt.Sprintf("%v/%v", groupMap[application.GroupID].FullName, application.Name)
		applicationMap[application.ID] = &ApplicationDetail{
			*applications[i],
			fullPath,
			fullName,
		}
	}

	return applicationMap, nil
}

func NewService(groupSvc groupservice.Service, manager *managerparam.Manager) Service {
	return &service{
		groupService:       groupSvc,
		applicationManager: manager.ApplicationManager,
	}
}
