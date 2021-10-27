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
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	ListByApplication(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
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

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	const op = "cluster manager: get by name"
	cluster, err := m.dao.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, _errCodeClusterNotFound)
		}
		return nil, errors.E(op, err)
	}
	return cluster, nil
}

func (m *manager) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	return m.dao.UpdateByID(ctx, id, cluster)
}

func (m *manager) ListByApplication(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	return m.dao.ListByApplication(ctx, applicationID)
}

func (m *manager) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	return m.dao.CheckClusterExists(ctx, cluster)
}
