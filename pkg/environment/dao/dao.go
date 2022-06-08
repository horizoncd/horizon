package dao

import (
	"context"
	"sort"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	appregionmodels "g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/environment/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	"gorm.io/gorm"
)

type DAO interface {
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
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) UpdateByID(ctx context.Context, id uint, environment *models.Environment) error {
	environmentInDB, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// set displayName
	environmentInDB.DisplayName = environment.DisplayName
	res := db.Save(&environmentInDB)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentInDB, res.Error.Error())
	}

	return nil
}

func (d *dao) CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(environment)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environment, result.Error
}

func (d *dao) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environments []*models.Environment

	result := db.Raw(common.EnvironmentListAll).Scan(&environments)

	if result.Error != nil {
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	sort.Sort(models.EnvironmentList(environments))
	return environments, nil
}

func (d *dao) DeleteByID(ctx context.Context, id uint) error {
	environment, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// remove related resources from different tables
	err = db.Transaction(func(tx *gorm.DB) error {
		// remove records from applicationRegion table
		res := tx.Where("environment_name = ?", environment.Name).Delete(&appregionmodels.ApplicationRegion{})
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, res.Error.Error())
		}

		// remove records from environmentRegion table
		res = tx.Where("environment_name = ?", environment.Name).Delete(&envregionmodels.EnvironmentRegion{})
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.RegionInDB, res.Error.Error())
		}

		// remove environment itself
		res = tx.Delete(&models.Environment{}, id)
		if res.Error != nil {
			return herrors.NewErrDeleteFailed(herrors.EnvironmentInDB, res.Error.Error())
		}
		return nil
	})

	return err
}

func (d *dao) GetByID(ctx context.Context, id uint) (*models.Environment, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environment models.Environment
	result := db.Raw(common.EnvironmentGetByID, id).First(&environment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentInDB, result.Error.Error())
	}

	return &environment, nil
}
