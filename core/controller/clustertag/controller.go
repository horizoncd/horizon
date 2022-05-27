package clustertag

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/clustertag/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, clusterID uint) (*ListResponse, error)
	Update(ctx context.Context, clusterID uint, r *UpdateRequest) error
}

type controller struct {
	clusterMgr     clustermanager.Manager
	clusterTagMgr  manager.Manager
	clusterGitRepo gitrepo.ClusterGitRepo
	applicationMgr appmanager.Manager
}

func NewController(clusterGitRepo gitrepo.ClusterGitRepo) Controller {
	return &controller{
		clusterMgr:     clustermanager.Mgr,
		clusterTagMgr:  manager.Mgr,
		clusterGitRepo: clusterGitRepo,
		applicationMgr: appmanager.Mgr,
	}
}

func (c *controller) List(ctx context.Context, clusterID uint) (_ *ListResponse, err error) {
	const op = "cluster tag controller: list"
	defer wlog.Start(ctx, op).StopPrint()

	clusterTags, err := c.clusterTagMgr.ListByClusterID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofClusterTags(clusterTags), nil
}

func (c *controller) Update(ctx context.Context, clusterID uint, r *UpdateRequest) (err error) {
	const op = "cluster tag controller: update"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	clusterTags := r.toClusterTags(clusterID, currentUser)

	if err := manager.ValidateUpsert(clusterTags); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return errors.E(op, err)
	}

	if err := c.clusterGitRepo.UpdateTags(ctx, application.Name, cluster.Name,
		cluster.Template, clusterTags); err != nil {
		if err != nil {
			return errors.E(op, err)
		}
	}

	if err := c.clusterTagMgr.UpsertByClusterID(ctx, clusterID, clusterTags); err != nil {
		return errors.E(op, err)
	}

	return nil
}
