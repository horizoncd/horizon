package templateschema

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/templateschema/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, clusterID uint) (*ListResponse, error)
	Update(ctx context.Context, clusterID uint, r *UpdateRequest) error
}

type controller struct {
	clusterMgr          clustermanager.Manager
	clusterSchemaTagMgr manager.Manager
}

func NewController() Controller {
	return &controller{
		clusterMgr:          clustermanager.Mgr,
		clusterSchemaTagMgr: manager.Mgr,
	}
}

func (c *controller) List(ctx context.Context, clusterID uint) (_ *ListResponse, err error) {
	const op = "cluster teplate scheme tag controller: list"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	clusterTags, err := c.clusterSchemaTagMgr.ListByClusterID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofClusterTags(clusterTags), nil
}

func (c *controller) Update(ctx context.Context, clusterID uint, r *UpdateRequest) (err error) {
	const op = "cluster template scheme tag controller: update"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	clusterTemplateSchemaTags := r.toClusterTags(clusterID, currentUser)

	if err := manager.ValidateUpsert(clusterTemplateSchemaTags); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}

	if err := c.clusterSchemaTagMgr.UpsertByClusterID(ctx, cluster.ID, clusterTemplateSchemaTags); err != nil {
		return errors.E(op, err)
	}

	return nil
}
