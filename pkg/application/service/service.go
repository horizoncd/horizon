package service

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/application/manager"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
)

var (
	Svc = NewService()
)

type Service interface {
	// GetByID get detail of an application by id
	GetByID(ctx context.Context, id uint) (*ApplicationDetail, error)
}

type service struct {
	groupService       groupservice.Service
	applicationManager manager.Manager
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

	applicationDetail := &ApplicationDetail{
		*application,
		fullPath,
	}
	return applicationDetail, nil
}

func NewService() Service {
	return &service{
		groupService:       groupservice.Svc,
		applicationManager: manager.Mgr,
	}
}
