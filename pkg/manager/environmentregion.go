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

func NewEnvironmentRegionManager(db *gorm.DB) EnvironmentRegionManager {
	return &environmentRegionManager{
		envRegionDAO: dao.NewEnvironmentRegionDAO(db),
		regionDAO:    dao.NewRegionDAO(db),
	}
}

type EnvironmentRegionManager interface {
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (
		*models.EnvironmentRegion, error)
	// ListByEnvironment list regions by env
	ListByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error)
	ListEnabledRegionsByEnvironment(ctx context.Context, env string) (models.RegionParts, error)
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	GetByEnvironmentAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	GetDefaultRegionByEnvironment(ctx context.Context, env string) (*models.EnvironmentRegion, error)
	SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error
	// ListAllEnvironmentRegions list all environmentRegions
	SetEnvironmentRegionIfAutoFree(ctx context.Context, id uint, autoFree bool) error
	ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	DeleteByID(ctx context.Context, id uint) error
}

type environmentRegionManager struct {
	envRegionDAO dao.EnvironmentRegionDAO
	regionDAO    dao.RegionDAO
}

// DeleteByID implements EnvironmentRegionManager
func (m *environmentRegionManager) DeleteByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.DeleteByID(ctx, id)
}

func (m *environmentRegionManager) GetDefaultRegionByEnvironment(ctx context.Context, env string) (
	*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetDefaultRegionByEnvironment(ctx, env)
}

func (m *environmentRegionManager) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.CreateEnvironmentRegion(ctx, er)
}

func (m *environmentRegionManager) ListByEnvironment(ctx context.Context,
	env string) ([]*models.EnvironmentRegion, error) {
	regions, err := m.envRegionDAO.ListRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func (m *environmentRegionManager) ListEnabledRegionsByEnvironment(ctx context.Context, env string) (
	models.RegionParts, error) {
	return m.envRegionDAO.ListEnabledRegionsByEnvironment(ctx, env)
}

func (m *environmentRegionManager) GetEnvironmentRegionByID(ctx context.Context,
	id uint) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByID(ctx, id)
}

func (m *environmentRegionManager) GetByEnvironmentAndRegion(ctx context.Context,
	env, region string) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByEnvAndRegion(ctx, env, region)
}

func (m *environmentRegionManager) SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.SetEnvironmentRegionToDefaultByID(ctx, id)
}

func (m *environmentRegionManager) SetEnvironmentRegionIfAutoFree(ctx context.Context, id uint, autoFree bool) error {
	return m.envRegionDAO.SetEnvironmentRegionIfAutoFree(ctx, id, autoFree)
}

func (m *environmentRegionManager) ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	return m.envRegionDAO.ListAllEnvironmentRegions(ctx)
}
