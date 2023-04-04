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

package dao

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/common"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ApplicationRegionDAO interface {
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error)
	ListByEnvApplicationID(ctx context.Context, env string, applicationID uint) (*models.ApplicationRegion, error)
	UpsertByApplicationID(ctx context.Context, applicationID uint, applicationRegions []*models.ApplicationRegion) error
}

type applicationRegionDAO struct {
	db *gorm.DB
}

func NewApplicationRegionDAO(db *gorm.DB) ApplicationRegionDAO {
	return &applicationRegionDAO{db: db}
}

func (d *applicationRegionDAO) ListByEnvApplicationID(ctx context.Context, env string,
	applicationID uint) (*models.ApplicationRegion, error) {
	var applicationRegion *models.ApplicationRegion
	result := d.db.WithContext(ctx).Raw(common.ApplicationRegionListByEnvApplicationID, env,
		applicationID).First(&applicationRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.ApplicationRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.ApplicationRegionInDB, result.Error.Error())
	}

	return applicationRegion, nil
}

func (d *applicationRegionDAO) ListByApplicationID(ctx context.Context,
	applicationID uint) ([]*models.ApplicationRegion, error) {
	var applicationRegions []*models.ApplicationRegion
	result := d.db.WithContext(ctx).Raw(common.ApplicationRegionListByApplicationID,
		applicationID).Scan(&applicationRegions)

	if result.Error != nil {
		return nil, perror.Wrapf(result.Error,
			"failed to list applicationRegions for applicationID: %d", applicationID)
	}

	return applicationRegions, nil
}

func (d *applicationRegionDAO) UpsertByApplicationID(ctx context.Context, applicationID uint,
	applicationRegions []*models.ApplicationRegion) error {
	var result *gorm.DB
	if len(applicationRegions) == 0 {
		result = d.db.WithContext(ctx).Exec(common.ApplicationRegionDeleteAllByApplicationID, applicationID)
		if result.Error != nil {
			return perror.Wrapf(result.Error,
				"failed to delete applicationRegions of applicationID: %d", applicationID)
		}
		return nil
	}

	result = d.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "application_id",
			}, {
				Name: "environment_name",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{"region_name", "updated_by"}),
	}).Create(applicationRegions)

	if result.Error != nil {
		return perror.Wrapf(result.Error,
			"failed to upsert applicationRegions of applicationID: %d", applicationID)
	}
	return nil
}
