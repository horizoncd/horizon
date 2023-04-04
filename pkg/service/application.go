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

type ApplicationService interface {
	// GetByID get application with full name and full path by id
	GetByID(ctx context.Context, id uint) (*models.ApplicationDetail, error)
	// GetByIDs get application map with full name and full path by ids
	GetByIDs(ctx context.Context, ids []uint) (map[uint]*models.ApplicationDetail, error)
}

type applicationService struct {
	groupSvc GroupService
	appMgr   manager.ApplicationManager
}

func NewApplicationService(groupSvc GroupService, manager *managerparam.Manager) ApplicationService {
	return &applicationService{
		groupSvc: groupSvc,
		appMgr:   manager.ApplicationManager,
	}
}

func (s applicationService) GetByID(ctx context.Context, id uint) (*models.ApplicationDetail, error) {
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

	appDetail := &models.ApplicationDetail{
		Application: *application,
		FullPath:    fullPath,
		FullName:    fullName,
	}
	return appDetail, nil
}

func (s applicationService) GetByIDs(ctx context.Context, ids []uint) (map[uint]*models.ApplicationDetail, error) {
	applicationMap := make(map[uint]*models.ApplicationDetail)
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

func (s applicationService) getAppsDetails(ctx context.Context,
	apps []*models.Application) ([]*models.ApplicationDetail, error) {
	groupIDs := make([]uint, len(apps))
	for i, app := range apps {
		groupIDs[i] = app.GroupID
	}
	groupMap, err := s.groupSvc.GetChildrenByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	appDetails := make([]*models.ApplicationDetail, len(apps))

	// 3. add full path and full name
	for i, app := range apps {
		fullPath := fmt.Sprintf("%v/%v", groupMap[app.GroupID].FullPath, app.Name)
		fullName := fmt.Sprintf("%v/%v", groupMap[app.GroupID].FullName, app.Name)
		appDetails[i] = &models.ApplicationDetail{
			Application: *app,
			FullPath:    fullPath,
			FullName:    fullName,
		}
	}

	return appDetails, nil
}
