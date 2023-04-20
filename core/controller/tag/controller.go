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

package tag

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	appmanager "github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	"github.com/horizoncd/horizon/pkg/param"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	"github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, resourceType string, resourceID uint) (*ListResponse, error)
	ListSubResourceTags(ctx context.Context, resourceType string,
		resourceID uint) (*ListResponse, error)
	Update(ctx context.Context, resourceType string, resourceID uint, r *UpdateRequest) error
}

type controller struct {
	clusterMgr     clustermanager.Manager
	tagMgr         tagmanager.Manager
	clusterGitRepo gitrepo.ClusterGitRepo
	applicationMgr appmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		clusterMgr:     param.ClusterMgr,
		tagMgr:         param.TagManager,
		clusterGitRepo: param.ClusterGitRepo,
		applicationMgr: param.ApplicationManager,
	}
}

func (c *controller) List(ctx context.Context, resourceType string, resourceID uint) (_ *ListResponse, err error) {
	const op = "cluster tag controller: list"
	defer wlog.Start(ctx, op).StopPrint()

	tags, err := c.tagMgr.ListByResourceTypeID(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	return ofTags(tags), nil
}

func (c *controller) Update(ctx context.Context, resourceType string, resourceID uint, r *UpdateRequest) (err error) {
	const op = "cluster tag controller: update"
	defer wlog.Start(ctx, op).StopPrint()

	tags := r.toTags(resourceType, resourceID)
	if err := tagmanager.ValidateUpsert(tags); err != nil {
		return err
	}

	if resourceType == common.ResourceCluster {
		cluster, err := c.clusterMgr.GetByID(ctx, resourceID)
		if err != nil {
			return err
		}
		application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
		if err != nil {
			return err
		}

		if err := c.clusterGitRepo.UpdateTags(ctx, application.Name, cluster.Name,
			cluster.Template, tags); err != nil {
			if err != nil {
				return err
			}
		}
	}

	return c.tagMgr.UpsertByResourceTypeID(ctx, resourceType, resourceID, tags)
}

func (c *controller) ListSubResourceTags(ctx context.Context, resourceType string,
	resourceID uint) (*ListResponse, error) {
	const op = "cluster tag controller: list sub resource tags"
	defer wlog.Start(ctx, op).StopPrint()

	var results []*models.Tag
	if resourceType == common.ResourceApplication {
		query := q.New(q.KeyWords{common.ParamApplicationID: resourceID})
		query.WithoutPagination = true
		_, clusters, err := c.clusterMgr.List(ctx, query)
		if err != nil {
			return nil, err
		}

		var clusterIDs []uint
		for _, cluster := range clusters {
			clusterIDs = append(clusterIDs, cluster.ID)
		}
		tags, err := c.tagMgr.ListByResourceTypeIDs(ctx, common.ResourceCluster, clusterIDs, true)
		if err != nil {
			return nil, err
		}
		results = append(results, tags...)
	}

	return ofTags(results), nil
}
