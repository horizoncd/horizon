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

	"github.com/horizoncd/horizon/core/common"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	appservice "github.com/horizoncd/horizon/pkg/application/service"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	tmodels "github.com/horizoncd/horizon/pkg/tag/models"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
)

type Service interface {
	// GetByID get detail of an application by id
	GetByID(ctx context.Context, id uint) (*ClusterDetail, error)
	// SyncDBWithGitRepo syncs template and tags in db when git repo files are updated
	SyncDBWithGitRepo(ctx context.Context, application *appmodels.Application,
		cluster *clustermodels.Cluster) (*clustermodels.Cluster, error)
}

type service struct {
	appSvc         appservice.Service
	clusterMgr     clustermanager.Manager
	trMgr          trmanager.Manager
	tagMgr         tagmanager.Manager
	clusterGitRepo gitrepo.ClusterGitRepo
}

var _ Service = (*service)(nil)

func NewService(applicationSvc appservice.Service, clusterGitRep gitrepo.ClusterGitRepo,
	manager *managerparam.Manager) Service {
	return &service{
		appSvc:         applicationSvc,
		clusterMgr:     manager.ClusterMgr,
		trMgr:          manager.TemplateReleaseMgr,
		tagMgr:         manager.TagMgr,
		clusterGitRepo: clusterGitRep,
	}
}

func (s service) GetByID(ctx context.Context, id uint) (*ClusterDetail, error) {
	cluster, err := s.clusterMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	application, err := s.appSvc.GetByID(ctx, cluster.ApplicationID)
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

func (s service) SyncDBWithGitRepo(ctx context.Context, application *appmodels.Application,
	cluster *clustermodels.Cluster) (*clustermodels.Cluster, error) {
	templateFromFile, err := s.clusterGitRepo.GetClusterTemplate(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}
	cluster.Template = templateFromFile.Name
	cluster.TemplateRelease = templateFromFile.Release
	cluster, err = s.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	files, err := s.clusterGitRepo.GetClusterValueFiles(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.FileName == common.GitopsFileTags {
			release, err := s.trMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
			if err != nil {
				return nil, err
			}
			midMap := file.Content[release.ChartName].(map[string]interface{})
			tagsMap := midMap[common.GitopsKeyTags].(map[string]interface{})
			tags := make([]*tmodels.TagBasic, 0, len(tagsMap))
			for k, v := range tagsMap {
				value, ok := v.(string)
				if !ok {
					continue
				}
				tags = append(tags, &tmodels.TagBasic{
					Key:   k,
					Value: value,
				})
			}
			return cluster, s.tagMgr.UpsertByResourceTypeID(ctx,
				common.ResourceCluster, cluster.ID, tags)
		}
	}
	return cluster, nil
}
