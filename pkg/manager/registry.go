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

	"github.com/horizoncd/horizon/pkg/dao"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type RegistryManager interface {
	// Create a registry
	Create(ctx context.Context, registry *models.Registry) (uint, error)
	// UpdateByID update a registry
	UpdateByID(ctx context.Context, id uint, registry *models.Registry) error
	// DeleteByID delete a registry by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Registry, error)
	// ListAll list all registries
	ListAll(ctx context.Context) ([]*models.Registry, error)
}

type registryManager struct {
	registryDAO dao.RegistryDAO
}

func NewRegistryManager(db *gorm.DB) RegistryManager {
	return &registryManager{
		registryDAO: dao.NewRegistryDAO(db),
	}
}

func (m registryManager) Create(ctx context.Context, registry *models.Registry) (uint, error) {
	return m.registryDAO.Create(ctx, registry)
}

func (m registryManager) GetByID(ctx context.Context, id uint) (*models.Registry, error) {
	return m.registryDAO.GetByID(ctx, id)
}

func (m registryManager) ListAll(ctx context.Context) ([]*models.Registry, error) {
	return m.registryDAO.ListAll(ctx)
}

func (m registryManager) UpdateByID(ctx context.Context, id uint, registry *models.Registry) error {
	return m.registryDAO.UpdateByID(ctx, id, registry)
}

func (m registryManager) DeleteByID(ctx context.Context, id uint) error {
	return m.registryDAO.DeleteByID(ctx, id)
}
