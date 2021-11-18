package clustertag

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/clustertag/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

var (
	Ctl = NewController()
)

type Controller interface {
	List(ctx context.Context, clusterID uint) (*ListResponse, error)
	Update(ctx context.Context, clusterID uint, r *UpdateRequest) error
}

type controller struct {
	clusterTagMgr manager.Manager
}

func NewController() Controller {
	return &controller{
		clusterTagMgr: manager.Mgr,
	}
}

func (c *controller) List(ctx context.Context, clusterID uint) (_ *ListResponse, err error) {
	const op = "cluster tag controller: list"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	clusterTags, err := c.clusterTagMgr.ListByClusterID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofClusterTags(clusterTags), nil
}

func (c *controller) Update(ctx context.Context, clusterID uint, r *UpdateRequest) (err error) {
	const op = "cluster tag controller: update"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	clusterTags := r.toClusterTags(clusterID, currentUser)

	if err := manager.ValidateUpsert(clusterTags); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}

	if err := c.clusterTagMgr.UpsertByClusterID(ctx, clusterID, clusterTags); err != nil {
		return errors.E(op, err)
	}

	return nil
}
