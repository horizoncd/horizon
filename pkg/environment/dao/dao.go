package dao

import (
	"context"
	"sort"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/environment/models"
)

type DAO interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*models.Environment, error)
	// UpdateByID update environment by id
	UpdateByID(ctx context.Context, id uint, environment *models.Environment) error
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) UpdateByID(ctx context.Context, id uint, environment *models.Environment) error {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// can only update displayName
	// check en exist
	var environmentInDB models.Environment
	res := db.Find(&environmentInDB, id)
	if res.RowsAffected == 0 {
		return herrors.NewErrNotFound(herrors.EnvironmentInDB, "rows affected = 0")
	}
	if res.Error != nil {
		return herrors.NewErrGetFailed(herrors.EnvironmentInDB, res.Error.Error())
	}

	// set displayName
	environmentInDB.DisplayName = environment.DisplayName
	res = db.Save(&environmentInDB)
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
