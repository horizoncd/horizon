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

type IDProviderManager interface {
	List(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
	Create(ctx context.Context, idp *models.IdentityProvider) (*models.IdentityProvider, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error)
	GetByCondition(ctx context.Context, condition q.Query) (*models.IdentityProvider, error)
	Update(ctx context.Context, id uint, param *models.IdentityProvider) (*models.IdentityProvider, error)
}

type idProviderManager struct {
	dao dao.IDProviderDAO
}

func NewIDProviderManager(db *gorm.DB) IDProviderManager {
	return &idProviderManager{
		dao: dao.NewIDProviderDAO(db),
	}
}

func (m *idProviderManager) List(ctx context.Context) ([]*models.IdentityProvider, error) {
	return m.dao.List(ctx)
}

func (m *idProviderManager) GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error) {
	return m.dao.GetProviderByName(ctx, name)
}

func (m *idProviderManager) Create(ctx context.Context,
	idp *models.IdentityProvider) (*models.IdentityProvider, error) {
	return m.dao.Create(ctx, idp)
}

func (m *idProviderManager) Delete(ctx context.Context, id uint) error {
	return m.dao.Delete(ctx, id)
}

func (m *idProviderManager) GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error) {
	return m.dao.GetByID(ctx, id)
}

func (m *idProviderManager) GetByCondition(ctx context.Context,
	condition q.Query) (*models.IdentityProvider, error) {
	return m.dao.GetByCondition(ctx, condition)
}

func (m *idProviderManager) Update(ctx context.Context,
	id uint, param *models.IdentityProvider) (*models.IdentityProvider, error) {
	return m.dao.Update(ctx, id, param)
}
