package service

import (
	"context"
	"fmt"
	"github.com/horizoncd/horizon/pkg/param/managerparam"

	"github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/application/models"
	groupsvc "github.com/horizoncd/horizon/pkg/group/service"
)

type Service interface {
	// GetByID get application with full name and full path by id
	GetByID(ctx context.Context, id uint) (*ApplicationDetail, error)
	// GetByIDs get application map with full name and full path by ids
	GetByIDs(ctx context.Context, ids []uint) (map[uint]*ApplicationDetail, error)
}

type service struct {
	groupSvc groupsvc.Service
	appMgr   manager.Manager
}

func NewService(groupSvc groupsvc.Service, manager *managerparam.Manager) Service {
	return &service{
		groupSvc: groupSvc,
		appMgr:   manager.ApplicationManager,
	}
}

func (s service) GetByID(ctx context.Context, id uint) (*ApplicationDetail, error) {
	application, err := s.appMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	group, err := s.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, application.Name)
	fullName := fmt.Sprintf("%v/%v", group.FullName, application.Name)

	appDetail := &ApplicationDetail{
		*application,
		fullPath,
		fullName,
	}
	return appDetail, nil
}

func (s service) GetByIDs(ctx context.Context, ids []uint) (map[uint]*ApplicationDetail, error) {
	applicationMap := make(map[uint]*ApplicationDetail)
	// 1. get applications
	applications, err := s.appMgr.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	appDetails, err := s.getAppsDetails(ctx, applications)
	if err != nil {
		return nil, err
	}

	for _, appDetail := range appDetails {
		applicationMap[appDetail.ID] = appDetail
	}

	return applicationMap, nil
}

func (s service) getAppsDetails(ctx context.Context, apps []*models.Application) ([]*ApplicationDetail, error) {
	groupIDs := make([]uint, len(apps))
	for i, app := range apps {
		groupIDs[i] = app.GroupID
	}
	groupMap, err := s.groupSvc.GetChildrenByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	appDetails := make([]*ApplicationDetail, len(apps))

	// 3. add full path and full name
	for i, app := range apps {
		fullPath := fmt.Sprintf("%v/%v", groupMap[app.GroupID].FullPath, app.Name)
		fullName := fmt.Sprintf("%v/%v", groupMap[app.GroupID].FullName, app.Name)
		appDetails[i] = &ApplicationDetail{
			Application: *app,
			FullPath:    fullPath,
			FullName:    fullName,
		}
	}

	return appDetails, nil
}
