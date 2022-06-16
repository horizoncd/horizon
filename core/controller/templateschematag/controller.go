package templateschematag

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/param"

	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	templateschematagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	List(ctx context.Context, clusterID uint) (*ListResponse, error)
	Update(ctx context.Context, clusterID uint, r *UpdateRequest) error
}

type controller struct {
	clusterMgr          clustermanager.Manager
	clusterSchemaTagMgr templateschematagmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		clusterMgr:          param.ClusterMgr,
		clusterSchemaTagMgr: param.ClusterSchemaTagMgr,
	}
}

func (c *controller) List(ctx context.Context, clusterID uint) (_ *ListResponse, err error) {
	const op = "cluster template scheme tag controller: list"
	defer wlog.Start(ctx, op).StopPrint()

	tags, err := c.clusterSchemaTagMgr.ListByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	return ofClusterTemplateSchemaTags(tags), nil
}

func (c *controller) Update(ctx context.Context, clusterID uint, r *UpdateRequest) (err error) {
	const op = "cluster template scheme tag controller: update"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	clusterTemplateSchemaTags := r.toClusterTemplateSchemaTags(clusterID, currentUser)

	if err := templateschematagmanager.ValidateUpsert(clusterTemplateSchemaTags); err != nil {
		return err
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}

	return c.clusterSchemaTagMgr.UpsertByClusterID(ctx, cluster.ID, clusterTemplateSchemaTags)
}
