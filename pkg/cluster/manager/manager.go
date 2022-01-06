package manager

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/dao"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

var (
	// Mgr is the global cluster manager
	Mgr = New()
)

const _errCodeClusterNotFound = errors.ErrorCode("ClusterNotFound")

type Manager interface {
	Create(ctx context.Context, cluster *models.Cluster,
		clusterTags []*clustertagmodels.ClusterTag, extraOwners []string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	ListByApplicationAndEnvs(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error)
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	ListByNameFuzzily(ctx context.Context, environment, name string, query *q.Query) (int,
		[]*models.ClusterWithEnvAndRegion, error)
	// ListUserAuthorizedClusterByNameFuzzily list cluster which is authorized to the specified user.
	// 1. name is the cluster's fuzzily name.
	// 2. applicationIDs is the applications' id which are authorized to the specified user.
	// 3. userInfo is the user id
	ListUserAuthorizedClusterByNameFuzzily(ctx context.Context,
		name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.Cluster, error)
}

func New() Manager {
	return &manager{
		dao:     dao.NewDAO(),
		userDAO: userdao.NewDAO(),
	}
}

type manager struct {
	dao     dao.DAO
	userDAO userdao.DAO
}

func (m *manager) Create(ctx context.Context, cluster *models.Cluster,
	clusterTags []*clustertagmodels.ClusterTag, extraOwners []string) (*models.Cluster, error) {
	users, err := m.userDAO.ListByEmail(ctx, extraOwners)
	if err != nil {
		return nil, err
	}

	return m.dao.Create(ctx, cluster, clusterTags, users)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	const op = "cluster manager: get by id"
	cluster, err := m.dao.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, _errCodeClusterNotFound)
		}
		return nil, errors.E(op, err)
	}
	return cluster, nil
}

func (m *manager) GetByName(ctx context.Context, clusterName string) (*models.Cluster, error) {
	return m.dao.GetByName(ctx, clusterName)
}

func (m *manager) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	return m.dao.UpdateByID(ctx, id, cluster)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}

func (m *manager) ListByApplicationAndEnvs(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	if query == nil {
		query = &q.Query{
			PageNumber: common.DefaultPageNumber,
			PageSize:   common.DefaultPageSize,
		}
	}
	if query.PageNumber < 1 {
		query.PageNumber = common.DefaultPageNumber
	}
	if query.PageSize < 1 {
		query.PageSize = common.DefaultPageSize
	}
	return m.dao.ListByApplicationAndEnvs(ctx, applicationID, environments, filter, query)
}

func (m *manager) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	return m.dao.ListByApplicationID(ctx, applicationID)
}

func (m *manager) ListByNameFuzzily(ctx context.Context, environment,
	name string, query *q.Query) (int, []*models.ClusterWithEnvAndRegion, error) {
	if query == nil {
		query = &q.Query{
			PageNumber: common.DefaultPageNumber,
			PageSize:   common.DefaultPageSize,
		}
	}
	if query.PageNumber < 1 {
		query.PageNumber = common.DefaultPageNumber
	}
	if query.PageSize < 1 {
		query.PageSize = common.DefaultPageSize
	}
	return m.dao.ListByNameFuzzily(ctx, environment, name, query)
}

func (m *manager) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	return m.dao.CheckClusterExists(ctx, cluster)
}

func (m *manager) ListUserAuthorizedClusterByNameFuzzily(ctx context.Context,
	name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.Cluster, error) {
	if query == nil {
		query = &q.Query{
			PageNumber: common.DefaultPageNumber,
			PageSize:   common.DefaultPageSize,
		}
	}
	if query.PageNumber < 1 {
		query.PageNumber = common.DefaultPageNumber
	}
	if query.PageSize < 1 {
		query.PageSize = common.DefaultPageSize
	}
	return m.dao.ListUserAuthorizedClusterByNameFuzzily(ctx, name, applicationIDs, userInfo, query)
}
