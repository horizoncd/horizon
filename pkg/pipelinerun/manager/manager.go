package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/pipelinerun/dao"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
)

var (
	// Mgr is the global pipelinerun manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetLastPipelinerunWithConfigCommit(ctx context.Context, clusterID uint) (*models.Pipelinerun, error)
}

type manager struct {
	dao dao.DAO
}

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

func (m *manager) Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error) {
	return m.dao.Create(ctx, pipelinerun)
}

func (m *manager) GetLastPipelinerunWithConfigCommit(ctx context.Context, clusterID uint) (*models.Pipelinerun, error) {
	return m.dao.GetLastPipelinerunWithConfigCommit(ctx, clusterID)
}
