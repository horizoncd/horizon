package slo

import (
	"context"

	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/param"
	pipelinemanager "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/manager"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/models"
)

type Controller interface {
	// GetApplicationPipelineStats return pipeline stats about an application
	GetApplicationPipelineStats(ctx context.Context, application, cluster string, pageNumber, pageSize int) (
		[]*models.PipelineStats, int64, error)
}

type controller struct {
	pipelinemanager    pipelinemanager.Manager
	applicationmanager applicationmanager.Manager
	clustermanager     clustermanager.Manager
}

func (c controller) GetApplicationPipelineStats(ctx context.Context, application, cluster string,
	pageNumber, pageSize int) ([]*models.PipelineStats, int64, error) {
	_, err := c.applicationmanager.GetByName(ctx, application)
	if err != nil {
		return nil, 0, err
	}
	if cluster != "" {
		_, err := c.clustermanager.GetByName(ctx, cluster)
		if err != nil {
			return nil, 0, err
		}
	}

	return c.pipelinemanager.ListPipelineStats(ctx, application, cluster, pageNumber, pageSize)
}

func NewController(param *param.Param) Controller {
	return &controller{
		pipelinemanager:    param.PipelineMgr,
		applicationmanager: param.ApplicationManager,
		clustermanager:     param.ClusterMgr,
	}
}
