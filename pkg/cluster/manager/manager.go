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
	"github.com/horizoncd/horizon/pkg/cluster/dao"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	userdao "github.com/horizoncd/horizon/pkg/user/dao"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/manager/manager.go -package=mock_manager
type Manager interface {
	Create(ctx context.Context, cluster *models.Cluster,
		tags []*tagmodels.Tag, extraMembers map[string]string) (*models.Cluster, error)
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
	cluster, err := m.dao.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (m *manager) GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Cluster, error) {
	cluster, err := m.dao.GetByID(ctx, id, true)
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

func (m *manager) CheckClusterExists(ctx context.Context, cluster string) (bool, error) {
	return m.dao.CheckClusterExists(ctx, cluster)
}

func (m *manager) List(ctx context.Context, query *q.Query, appIDs ...uint) (int, []*models.ClusterWithRegion, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	currentUserID := currentUser.GetID()
	return m.dao.List(ctx, query, currentUserID, true, appIDs...)
}

func (m *manager) GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*models.Cluster, error) {
	return m.dao.GetByNameFuzzily(ctx, name, true)
}

func (m *manager) ListByApplicationID(ctx context.Context,
	applicationID uint) (int, []*models.ClusterWithRegion, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	currentUserID := currentUser.GetID()
	return m.dao.List(ctx, q.New(q.KeyWords{common.ParamApplicationID: applicationID}), currentUserID, false)
}

func (m *manager) ListClusterWithExpiry(ctx context.Context,
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
