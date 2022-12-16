package service

import (
	"context"
	"fmt"

	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/application/models"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
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

	appDetails, err := s.GetAppsDetails(ctx, applications)
	if err != nil {
		return nil, err
	}

	for _, appDetail := range appDetails {
		applicationMap[appDetail.ID] = appDetail
	}

	return applicationMap, nil
}

func (s service) GetAppsDetails(ctx context.Context, apps []*models.Application) ([]*ApplicationDetail, error) {
	var groupIDs []uint
	for _, application := range apps {
		groupIDs = append(groupIDs, application.GroupID)
	}
	groupMap, err := s.groupService.GetChildrenByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	appDetails := make([]*ApplicationDetail, 0, len(apps))

	// 3. add full path and full name
	for i, application := range apps {
		fullPath := fmt.Sprintf("%v/%v", groupMap[application.GroupID].FullPath, application.Name)
		fullName := fmt.Sprintf("%v/%v", groupMap[application.GroupID].FullName, application.Name)
		appDetails = append(appDetails, &ApplicationDetail{
			Application: *apps[i],
			FullPath:    fullPath,
			FullName:    fullName,
		})
	}

	return appDetails, nil
}

func NewService(groupSvc groupservice.Service, manager *managerparam.Manager) Service {
	return &service{
		groupService:       groupSvc,
		applicationManager: manager.ApplicationManager,
	}
}
