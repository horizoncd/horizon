// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/cluster/manager/manager.go -package=mock_manager
type ClusterManager interface {
	Create(ctx context.Context, cluster *models.Cluster,
		tags []*models.Tag, extraMembers map[string]string) (*models.Cluster, error)
	GetByID(ctx context.Context, id uint) (*models.Cluster, error)
	GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Cluster, error)
	GetByName(ctx context.Context, clusterName string) (*models.Cluster, error)
	UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error)
	DeleteByID(ctx context.Context, id uint) error
	CheckClusterExists(ctx context.Context, cluster string) (bool, error)
	List(ctx context.Context, query *q.Query, appIDs ...uint) (int, []*models.ClusterWithRegion, error)
	ListByApplicationID(ctx context.Context, applicationID uint) (int, []*models.ClusterWithRegion, error)
	ListClusterWithExpiry(ctx context.Context, query *q.Query) ([]*models.Cluster, error)
	GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*models.Cluster, error)
}

func NewClusterManager(db *gorm.DB) ClusterManager {
	return &clusterManager{
		dao:     dao.NewClusterDAO(db),
		userDAO: dao.NewUserDAO(db),
	}
}

type clusterManager struct {
	dao     dao.ClusterDAO
	userDAO dao.UserDAO
}

func (m *clusterManager) Create(ctx context.Context, cluster *models.Cluster,
	tags []*models.Tag, extraMembers map[string]string) (*models.Cluster, error) {
	emails := make([]string, 0, len(extraMembers))
	for email := range extraMembers {
		emails = append(emails, email)
	}
	users, err := m.userDAO.ListByEmail(ctx, emails)
	if err != nil {
		return nil, err
	}
	extraMembersWithUser := make(map[*models.User]string)
	for _, user := range users {
		extraMembersWithUser[user] = extraMembers[user.Email]
	}

	return m.dao.Create(ctx, cluster, tags, extraMembersWithUser)
}

func (m *clusterManager) GetByID(ctx context.Context, id uint) (*models.Cluster, error) {
	cluster, err := m.dao.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (m *clusterManager) GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Cluster, error) {
	cluster, err := m.dao.GetByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (m *clusterManager) GetByName(ctx context.Context, clusterName string) (*models.Cluster, error) {
	return m.dao.GetByName(ctx, clusterName)
}

func (m *clusterManager) UpdateByID(ctx context.Context, id uint, cluster *models.Cluster) (*models.Cluster, error) {
	return m.dao.UpdateByID(ctx, id, cluster)
}

func (m *clusterManager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}

func (m *clusterManager) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	return m.dao.CheckClusterExists(ctx, cluster)
}

func (m *clusterManager) List(ctx context.Context, query *q.Query,
	appIDs ...uint) (int, []*models.ClusterWithRegion, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	currentUserID := currentUser.GetID()
	return m.dao.List(ctx, query, currentUserID, true, appIDs...)
}

func (m *clusterManager) GetByNameFuzzilyIncludeSoftDelete(ctx context.Context,
	name string) ([]*models.Cluster, error) {
	return m.dao.GetByNameFuzzily(ctx, name, true)
}

func (m *clusterManager) ListByApplicationID(ctx context.Context,
	applicationID uint) (int, []*models.ClusterWithRegion, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	currentUserID := currentUser.GetID()
	return m.dao.List(ctx, q.New(q.KeyWords{common.ParamApplicationID: applicationID}), currentUserID, false)
}

func (m *clusterManager) ListClusterWithExpiry(ctx context.Context,
	query *q.Query) ([]*models.Cluster, error) {
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
	return m.dao.ListClusterWithExpiry(ctx, query)
}
