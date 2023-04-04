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

type ApplicationRegionManager interface {
	// ListByEnvApplicationID list applicationRegion by env and applicationID
	ListByEnvApplicationID(ctx context.Context, env string, applicationID uint) (*models.ApplicationRegion, error)
	// ListByApplicationID list applicationRegions by applicationID
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error)
	// UpsertByApplicationID upsert application regions
	UpsertByApplicationID(ctx context.Context, applicationID uint, applicationRegions []*models.ApplicationRegion) error
}

func NewApplicationRegionManager(db *gorm.DB) ApplicationRegionManager {
	return &applicationRegionManager{
		dao: dao.NewApplicationRegionDAO(db),
	}
}

type applicationRegionManager struct {
	dao dao.ApplicationRegionDAO
}

func (m *applicationRegionManager) ListByEnvApplicationID(ctx context.Context, env string,
	applicationID uint) (*models.ApplicationRegion, error) {
	return m.dao.ListByEnvApplicationID(ctx, env, applicationID)
}

func (m *applicationRegionManager) ListByApplicationID(ctx context.Context,
	applicationID uint) ([]*models.ApplicationRegion, error) {
	return m.dao.ListByApplicationID(ctx, applicationID)
}

func (m *applicationRegionManager) UpsertByApplicationID(ctx context.Context,
	applicationID uint, applicationRegions []*models.ApplicationRegion) error {
	return m.dao.UpsertByApplicationID(ctx, applicationID, applicationRegions)
}
