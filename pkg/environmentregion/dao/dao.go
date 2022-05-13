package dao

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"gorm.io/gorm"
)

type DAO interface {
	// GetEnvironmentRegionByID ...
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	// GetEnvironmentRegionByEnvAndRegion get
	GetEnvironmentRegionByEnvAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (*models.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by env
	ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error)
	// ListAllEnvironmentRegions list all environmentRegions
	ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	// UpdateEnvironmentRegionByID update environmentRegion by id
	UpdateEnvironmentRegionByID(ctx context.Context, id uint, environmentRegion *models.EnvironmentRegion) error
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environmentRegions []*models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionGetByEnvAndRegion, er.EnvironmentName,
		er.RegionName).Scan(&environmentRegions)
	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	if len(environmentRegions) > 0 {
		return nil, perror.Wrap(herrors.ErrPairConflict, "environmentRegion pair already exists")
	}

	result = db.Create(er)
	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	return er, result.Error
}

func (d *dao) ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var regions []string
	result := db.Raw(common.EnvironmentListRegion, env).Scan(&regions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return regions, result.Error
}

func (d *dao) GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environmentRegion models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionGetByID, id).First(&environmentRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return &environmentRegion, result.Error
}

func (d *dao) GetEnvironmentRegionByEnvAndRegion(ctx context.Context,
	env, region string) (*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environmentRegion models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionGet, env, region).First(&environmentRegion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentRegionInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	return &environmentRegion, nil
}

func (d *dao) UpdateEnvironmentRegionByID(ctx context.Context, id uint,
	environmentRegion *models.EnvironmentRegion) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	var environmentRegions []*models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionGetByEnvAndRegion, environmentRegion.EnvironmentName,
		environmentRegion.RegionName).Scan(&environmentRegions)
	if result.Error != nil {
		return herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	if len(environmentRegions) > 1 {
		return perror.Wrap(herrors.ErrPairConflict, "more than one environmentRegion pair exists")
	}
	if len(environmentRegions) == 1 && environmentRegions[0].ID != id {
		return perror.Wrap(herrors.ErrPairConflict, "environmentRegion pair already exists")
	}

	result = db.Exec(common.EnvironmentRegionUpdateByID, environmentRegion.EnvironmentName,
		environmentRegion.RegionName, id)
	if result.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentRegionInDB, "num of updated rows is zero")
	}
	return nil
}

func (d *dao) ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environmentRegions []*models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionListAll).Scan(&environmentRegions)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environmentRegions, nil
}
