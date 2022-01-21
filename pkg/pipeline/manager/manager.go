package manager

import (
	"context"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
	"g.hz.netease.com/horizon/pkg/pipeline/dao"
	"g.hz.netease.com/horizon/pkg/pipeline/models"
)

var (
	// Mgr is the global pipeline manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, results *metrics.PipelineResults) error
	ListPipelineSLOsByEnvsAndTimeRange(ctx context.Context, envs []string, start, end int64) ([]*models.PipelineSLO, error)
}

type manager struct {
	dao dao.DAO
}

func (m manager) ListPipelineSLOsByEnvsAndTimeRange(ctx context.Context, envs []string,
	start, end int64) ([]*models.PipelineSLO, error) {
	return m.dao.ListPipelineSLOsByEnvsAndTimeRange(ctx, envs, start, end)
}

func (m manager) Create(ctx context.Context, results *metrics.PipelineResults) error {
	return m.dao.Create(ctx, results)
}

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}
