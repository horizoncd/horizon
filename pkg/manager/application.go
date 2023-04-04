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

	"github.com/horizoncd/horizon/lib/q"
	applicationdao "github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/application/manager/manager.go -package=mock_manager
type ApplicationManager interface {
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*models.Application, error)
	Create(ctx context.Context, application *models.Application,
		extraMembers map[string]string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
	Transfer(ctx context.Context, id uint, groupID uint) error
	List(ctx context.Context, groupIDs []uint, query *q.Query) (int, []*models.Application, error)
}

func NewApplicationManager(db *gorm.DB) ApplicationManager {
	return &applicationManager{
		applicationDAO: applicationdao.NewApplicationDAO(db),
		groupDAO:       applicationdao.NewGroupDAO(db),
		userDAO:        applicationdao.NewUserDAO(db),
	}
}

type applicationManager struct {
	applicationDAO applicationdao.ApplicationDAO
	groupDAO       applicationdao.GroupDAO
	userDAO        applicationdao.UserDAO
}

func (m *applicationManager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error) {
	return m.applicationDAO.GetByNameFuzzily(ctx, name, false)
}

func (m *applicationManager) GetByNameFuzzilyIncludeSoftDelete(ctx context.Context,
	name string) ([]*models.Application, error) {
	return m.applicationDAO.GetByNameFuzzily(ctx, name, true)
}

func (m *applicationManager) GetByID(ctx context.Context, id uint) (*models.Application, error) {
	application, err := m.applicationDAO.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *applicationManager) GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Application, error) {
	application, err := m.applicationDAO.GetByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *applicationManager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error) {
	applications, err := m.applicationDAO.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (m *applicationManager) GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error) {
	return m.applicationDAO.GetByGroupIDs(ctx, groupIDs)
}

func (m *applicationManager) GetByName(ctx context.Context, name string) (*models.Application, error) {
	application, err := m.applicationDAO.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *applicationManager) Create(ctx context.Context, application *models.Application,
	extraMembers map[string]string) (*models.Application, error) {
	emails := make([]string, 0, len(extraMembers))
	for k := range extraMembers {
		emails = append(emails, k)
	}

	users, err := m.userDAO.ListByEmail(ctx, emails)
	if err != nil {
		return nil, err
	}

	extraMembersWithUser := make(map[*models.User]string)
	for _, user := range users {
		extraMembersWithUser[user] = extraMembers[user.Email]
	}

	return m.applicationDAO.Create(ctx, application, extraMembersWithUser)
}

func (m *applicationManager) UpdateByID(ctx context.Context,
	id uint, application *models.Application) (*models.Application, error) {
	return m.applicationDAO.UpdateByID(ctx, id, application)
}

func (m *applicationManager) DeleteByID(ctx context.Context, id uint) error {
	return m.applicationDAO.DeleteByID(ctx, id)
}

func (m *applicationManager) Transfer(ctx context.Context, id uint, groupID uint) error {
	return m.applicationDAO.TransferByID(ctx, id, groupID)
}

func (m *applicationManager) List(ctx context.Context,
	groupIDs []uint, query *q.Query) (int, []*models.Application, error) {
	return m.applicationDAO.List(ctx, groupIDs, query)
}
