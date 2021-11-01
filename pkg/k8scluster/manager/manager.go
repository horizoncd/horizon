package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/k8scluster/dao"
	"g.hz.netease.com/horizon/pkg/k8scluster/models"
)

var (
	// Mgr is the global k8sCluster manager
	Mgr = New()
)

type Manager interface {
	// Create a k8sCluster
	Create(ctx context.Context, k8sCluster *models.K8SCluster) (*models.K8SCluster, error)
	// ListAll list all k8sClusters
	ListAll(ctx context.Context) ([]*models.K8SCluster, error)
}

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) ListAll(ctx context.Context) ([]*models.K8SCluster, error) {
	return m.dao.ListAll(ctx)
}

func (m *manager) Create(ctx context.Context, k8sCluster *models.K8SCluster) (*models.K8SCluster, error) {
	return m.dao.Create(ctx, k8sCluster)
}
