// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"fmt"

	"github.com/horizoncd/horizon/pkg/manager"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
)

type ClusterService interface {
	// GetByID get detail of an application by id
	GetByID(ctx context.Context, id uint) (*models.ClusterDetail, error)
}

type clusterService struct {
	applicationService ApplicationService
	clusterManager     manager.ClusterManager
}

func (s clusterService) GetByID(ctx context.Context, id uint) (*models.ClusterDetail, error) {
	cluster, err := s.clusterManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	application, err := s.applicationService.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", application.FullPath, cluster.Name)

	clusterDetail := &models.ClusterDetail{
		Cluster:  *cluster,
		FullPath: fullPath,
	}
	return clusterDetail, nil
}

func NewClusterService(applicationSvc ApplicationService, manager *managerparam.Manager) ClusterService {
	return &clusterService{
		applicationService: applicationSvc,
		clusterManager:     manager.ClusterMgr,
	}
}
