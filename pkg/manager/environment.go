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

func NewEnvironmentManager(db *gorm.DB) EnvironmentManager {
	return &environmentManager{
		envDAO:    dao.NewEnvironmentDAO(db),
		regionDAO: dao.NewRegionDAO(db),
	}
}

type EnvironmentManager interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*models.Environment, error)
	// UpdateByID update environment by id
	UpdateByID(ctx context.Context, id uint, environment *models.Environment) error
	// DeleteByID delete environment by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get environment by id
	GetByID(ctx context.Context, id uint) (*models.Environment, error)
	// GetByName get environment by name
	GetByName(ctx context.Context, name string) (*models.Environment, error)
}

type environmentManager struct {
	envDAO    dao.EnvironmentDAO
	regionDAO dao.RegionDAO
}

func (m *environmentManager) GetByID(ctx context.Context, id uint) (*models.Environment, error) {
	return m.envDAO.GetByID(ctx, id)
}

func (m *environmentManager) GetByName(ctx context.Context, name string) (*models.Environment, error) {
	return m.envDAO.GetByName(ctx, name)
}

func (m *environmentManager) DeleteByID(ctx context.Context, id uint) error {
	return m.envDAO.DeleteByID(ctx, id)
}

func (m *environmentManager) UpdateByID(ctx context.Context, id uint, environment *models.Environment) error {
	return m.envDAO.UpdateByID(ctx, id, environment)
}

func (m *environmentManager) CreateEnvironment(ctx context.Context,
	environment *models.Environment) (*models.Environment, error) {
	return m.envDAO.CreateEnvironment(ctx, environment)
}

func (m *environmentManager) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	return m.envDAO.ListAllEnvironment(ctx)
}
