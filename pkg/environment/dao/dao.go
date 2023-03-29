package dao

import (
	"context"
	"sort"

	appregionmodels "github.com/horizoncd/horizon/pkg/applicationregion/models"
	"github.com/horizoncd/horizon/pkg/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/pkg/environment/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
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
	// GetByName get environment by name
	GetByName(ctx context.Context, name string) (*models.Environment, error)
}

type dao struct{ db *gorm.DB }

// NewDAO returns an instance of the default DAO
func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

func (d *dao) UpdateByID(ctx context.Context, id uint, environment *models.Environment) error {
	environmentInDB, err := d.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// set displayName and autoFree
	environmentInDB.DisplayName = environment.DisplayName
	res := d.db.WithContext(ctx).Save(&environmentInDB)
	if res.Error != nil {
		return herrors.NewErrUpdateFailed(herrors.EnvironmentInDB, res.Error.Error())
	}

	return nil
}

func (d *dao) CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	result := d.db.WithContext(ctx).Create(environment)

	if result.Error != nil {
		return nil, herrors.NewErrInsertFailed(herrors.EnvironmentRegionInDB, result.Error.Error())
	}

	return environment, result.Error
}

func (d *dao) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	var environments []*models.Environment

	result := d.db.WithContext(ctx).Raw(common.EnvironmentListAll).Scan(&environments)

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

	// remove related resources from different tables
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	var environment models.Environment
	result := d.db.WithContext(ctx).Raw(common.EnvironmentGetByID, id).First(&environment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentInDB, result.Error.Error())
	}

	return &environment, nil
}

func (d *dao) GetByName(ctx context.Context, name string) (*models.Environment, error) {
	var environment models.Environment
	result := d.db.WithContext(ctx).Raw(common.EnvironmentGetByName, name).First(&environment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, herrors.NewErrNotFound(herrors.EnvironmentInDB, result.Error.Error())
		}
		return nil, herrors.NewErrGetFailed(herrors.EnvironmentInDB, result.Error.Error())
	}

	return &environment, nil
}
