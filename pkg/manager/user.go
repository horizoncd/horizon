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
	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/user/manager/manager_mock.go -package=mock_manager
type UserManager interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	List(ctx context.Context, query *q.Query) (int64, []*models.User, error)
	GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetUserByIDs(ctx context.Context, userIDs []uint) ([]*models.User, error)
	GetUserMapByIDs(ctx context.Context, userIDs []uint) (map[uint]*models.User, error)
	ListByEmail(ctx context.Context, emails []string) ([]*models.User, error)
	UpdateByID(ctx context.Context, id uint, db *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
}

type userManager struct {
	dao dao.UserDAO
}

func NewUserManager(db *gorm.DB) UserManager {
	return &userManager{dao: dao.NewUserDAO(db)}
}

func (m *userManager) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return m.dao.Create(ctx, user)
}

func (m *userManager) List(ctx context.Context, query *q.Query) (int64, []*models.User, error) {
	return m.dao.List(ctx, query)
}

func (m *userManager) ListByEmail(ctx context.Context, emails []string) ([]*models.User, error) {
	return m.dao.ListByEmail(ctx, emails)
}

func (m *userManager) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	return m.dao.GetByID(ctx, userID)
}

func (m *userManager) GetUserByIDs(ctx context.Context, userIDs []uint) ([]*models.User, error) {
	return m.dao.GetByIDs(ctx, userIDs)
}

func (m *userManager) GetUserMapByIDs(ctx context.Context, userIDs []uint) (map[uint]*models.User, error) {
	users, err := m.GetUserByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint]*models.User)
	for _, user := range users {
		tmp := user
		userMap[user.ID] = tmp
	}
	return userMap, nil
}

func (m *userManager) GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error) {
	return m.dao.GetUserByIDP(ctx, email, idp)
}

func (m *userManager) UpdateByID(ctx context.Context, id uint, db *models.User) (*models.User, error) {
	return m.dao.UpdateByID(ctx, id, db)
}

func (m *userManager) DeleteUser(ctx context.Context, id uint) error {
	return m.dao.DeleteUser(ctx, id)
}
