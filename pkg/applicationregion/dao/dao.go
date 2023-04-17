package dao

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/applicationregion/models"
	"github.com/horizoncd/horizon/pkg/common"
	perror "github.com/horizoncd/horizon/pkg/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DAO interface {
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error)
	ListByEnvApplicationID(ctx context.Context, env string, applicationID uint) (*models.ApplicationRegion, error)
	UpsertByApplicationID(ctx context.Context, applicationID uint, applicationRegions []*models.ApplicationRegion) error
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) ListByEnvApplicationID(ctx context.Context, env string,
	applicationID uint,
) (*models.ApplicationRegion, error) {
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

func (d *dao) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error) {
	var applicationRegions []*models.ApplicationRegion
	result := d.db.WithContext(ctx).Raw(common.ApplicationRegionListByApplicationID,
		applicationID).Scan(&applicationRegions)

	if result.Error != nil {
		return nil, perror.Wrapf(result.Error,
			"failed to list applicationRegions for applicationID: %d", applicationID)
	}

	return applicationRegions, nil
}

func (d *dao) UpsertByApplicationID(ctx context.Context, applicationID uint,
	applicationRegions []*models.ApplicationRegion,
) error {
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
