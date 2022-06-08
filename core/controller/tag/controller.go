package tag

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/tag/manager"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, resourceType string, resourceID uint) (*ListResponse, error)
	Update(ctx context.Context, resourceType string, resourceID uint, r *UpdateRequest) error
}

type controller struct {
	clusterMgr     clustermanager.Manager
	tagMgr         manager.Manager
	clusterGitRepo gitrepo.ClusterGitRepo
	applicationMgr appmanager.Manager
}

func NewController(clusterGitRepo gitrepo.ClusterGitRepo) Controller {
	return &controller{
		clusterMgr:     clustermanager.Mgr,
		tagMgr:         manager.Mgr,
		clusterGitRepo: clusterGitRepo,
		applicationMgr: appmanager.Mgr,
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

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	tags := r.toTags(resourceType, resourceID, currentUser)

	if err := manager.ValidateUpsert(tags); err != nil {
		return err
	}

	if resourceType == models.TypeCluster {
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
