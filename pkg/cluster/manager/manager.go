package manager

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/dao"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/manager/manager.go -package=mock_manager
type Manager interface {
	Create(ctx context.Context, cluster *models.Cluster,
		tags []*tagmodels.Tag, extraMembers map[string]string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	ListByApplicationEnvsTags(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query, ts []tagmodels.TagSelector) (int, []*models.ClusterWithRegion, error)
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error)
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	ListByNameFuzzily(ctx context.Context, environment, name string, query *q.Query) (int,
		[]*models.ClusterWithRegion, error)
	// ListUserAuthorizedByNameFuzzily list cluster which is authorized to the specified user.
	// 1. name is the cluster's fuzzily name.
	// 2. applicationIDs is the applications' id which are authorized to the specified user.
	// 3. userInfo is the user id
	ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
		name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithRegion, error)
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao:     dao.NewDAO(db),
		userDAO: userdao.NewDAO(db),
	}
}

type manager struct {
	dao     dao.DAO
	userDAO userdao.DAO
}

func (m *manager) Create(ctx context.Context, cluster *models.Cluster,
	tags []*tagmodels.Tag, extraMembers map[string]string) (*models.Cluster, error) {
	emails := make([]string, 0, len(extraMembers))
	for email := range extraMembers {
		emails = append(emails, email)
	}
	users, err := m.userDAO.ListByEmail(ctx, emails)
	if err != nil {
		return nil, err
	}
	extraMembersWithUser := make(map[*usermodels.User]string)
	for _, user := range users {
		extraMembersWithUser[user] = extraMembers[user.Email]
	}

	return m.dao.Create(ctx, cluster, tags, extraMembersWithUser)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	cluster, err := m.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
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

func (m *manager) ListByApplicationEnvsTags(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query, ts []tagmodels.TagSelector) (int, []*models.ClusterWithRegion, error) {
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
	return m.dao.ListByApplicationEnvsTags(ctx, applicationID, environments, filter, query, ts)
}

func (m *manager) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.Cluster, error) {
	return m.dao.ListByApplicationID(ctx, applicationID)
}

func (m *manager) ListByNameFuzzily(ctx context.Context, environment,
	name string, query *q.Query) (int, []*models.ClusterWithRegion, error) {
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

func (m *manager) ListUserAuthorizedByNameFuzzily(ctx context.Context, environment,
	name string, applicationIDs []uint, userInfo uint, query *q.Query) (int, []*models.ClusterWithRegion, error) {
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
	return m.dao.ListUserAuthorizedByNameFuzzily(ctx, environment, name, applicationIDs, userInfo, query)
}
