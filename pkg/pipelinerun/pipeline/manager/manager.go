package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/dao"
	"gorm.io/gorm"
)

type Manager interface {
	Create(ctx context.Context, results *metrics.PipelineResults) error
	DeleteByClusterName(ctx context.Context, clusterName string) error
}

type manager struct {
	dao dao.DAO
}

func (m manager) Create(ctx context.Context, results *metrics.PipelineResults) error {
	return m.dao.Create(ctx, results)
}

func (m manager) DeleteByClusterName(ctx context.Context, clusterName string) error {
	return m.dao.DeleteByClusterName(ctx, clusterName)
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}
