package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/common"
	perrors "g.hz.netease.com/horizon/pkg/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DAO interface {
	ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error)
	UpsertByApplicationID(ctx context.Context, applicationID uint, applicationRegions []*models.ApplicationRegion) error
}

type dao struct {
}

func NewDAO() DAO {
	return &dao{}
}

func (d *dao) ListByApplicationID(ctx context.Context, applicationID uint) ([]*models.ApplicationRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var applicationRegions []*models.ApplicationRegion
	result := db.Raw(common.ApplicationRegionListByApplicationID, applicationID).Scan(&applicationRegions)

	if result.Error != nil {
		return nil, perrors.Wrapf(result.Error,
			"failed to list applicationRegions for applicationID: %d", applicationID)
	}

	return applicationRegions, nil
}

func (d *dao) UpsertByApplicationID(ctx context.Context, applicationID uint,
	applicationRegions []*models.ApplicationRegion) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	var result *gorm.DB
	if len(applicationRegions) == 0 {
		result = db.Exec(common.ApplicationRegionDeleteAllByApplicationID, applicationID)
		if result.Error != nil {
			return perrors.Wrapf(result.Error,
				"failed to delete applicationRegions of applicationID: %d", applicationID)
		}
		return nil
	}

	result = db.Clauses(clause.OnConflict{
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
		return perrors.Wrapf(result.Error,
			"failed to upsert applicationRegions of applicationID: %d", applicationID)
	}
	return nil
}
