package manager

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/pkg/cluster/dao"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

var (
	// Mgr is the global cluster manager
	Mgr = New()
)

const _errCodeClusterNotFound = errors.ErrorCode("ClusterNotFound")

type Manager interface {
	Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error)
	GetByName(ctx context.Context, name string) (*models.Cluster, error)
	ListByApplication(ctx context.Context, application string) ([]*models.Cluster, error)
}

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, error) {
	return m.dao.Create(ctx, cluster)
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Cluster, error) {
	const op = "cluster manager: get by name"
	cluster, err := m.dao.GetByName(ctx, name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, _errCodeClusterNotFound)
		}
		return nil, errors.E(op, err)
	}
	return cluster, nil
}

func (m *manager) ListByApplication(ctx context.Context, application string) ([]*models.Cluster, error) {
	return m.dao.ListByApplication(ctx, application)
}
