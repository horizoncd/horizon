package manager

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/pipelinerun/dao"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
)

var (
	// Mgr is the global pipelinerun manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, pipelinerun *models.Pipelinerun) (*models.Pipelinerun, error)
	GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error)
	GetByClusterID(ctx context.Context, clusterID uint, canRollback bool,
		query q.Query) (int, []*models.Pipelinerun, error)
	DeleteByID(ctx context.Context, pipelinerunID uint) error
	UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error
	GetLatestByClusterIDAndAction(ctx context.Context, clusterID uint, action string) (*models.Pipelinerun, error)
	UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error
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

func (m *manager) GetByID(ctx context.Context, pipelinerunID uint) (*models.Pipelinerun, error) {
	return m.dao.GetByID(ctx, pipelinerunID)
}

func (m *manager) DeleteByID(ctx context.Context, pipelinerunID uint) error {
	return m.dao.DeleteByID(ctx, pipelinerunID)
}

func (m *manager) UpdateConfigCommitByID(ctx context.Context, pipelinerunID uint, commit string) error {
	return m.dao.UpdateConfigCommitByID(ctx, pipelinerunID, commit)
}

func (m *manager) GetLatestByClusterIDAndAction(ctx context.Context,
	clusterID uint, action string) (*models.Pipelinerun, error) {
	return m.dao.GetLatestByClusterIDAndAction(ctx, clusterID, action)
}

func (m *manager) UpdateResultByID(ctx context.Context, pipelinerunID uint, result *models.Result) error {
	return m.dao.UpdateResultByID(ctx, pipelinerunID, result)
}

func (m *manager) GetByClusterID(ctx context.Context,
	clusterID uint, canRollback bool, query q.Query) (int, []*models.Pipelinerun, error) {
	return m.dao.GetByClusterID(ctx, clusterID, canRollback, query)
}
