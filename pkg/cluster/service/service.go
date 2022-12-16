package service

import (
	"context"
	"fmt"

	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
)

type Service interface {
	// GetByID get detail of an application by id
	GetByID(ctx context.Context, id uint) (*ClusterDetail, error)
}

type service struct {
	applicationService applicationservice.Service
	clusterManager     clustermanager.Manager
}

func (s service) GetByID(ctx context.Context, id uint) (*ClusterDetail, error) {
	cluster, err := s.clusterManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	application, err := s.applicationService.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", application.FullPath, cluster.Name)

	clusterDetail := &ClusterDetail{
		*cluster,
		fullPath,
	}
	return clusterDetail, nil
}

func NewService(applicationSvc applicationservice.Service, manager *managerparam.Manager) Service {
	return &service{
		applicationService: applicationSvc,
		clusterManager:     manager.ClusterMgr,
	}
}
