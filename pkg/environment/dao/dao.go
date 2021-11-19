package dao

import (
	"context"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/common"
	"g.hz.netease.com/horizon/pkg/environment/models"
)

type DAO interface {
	EnvironmentDAO
	EnvironmentRegionDAO
}

type EnvironmentDAO interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*models.Environment, error)
}

type EnvironmentRegionDAO interface {
	// GetEnvironmentRegionByID ...
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	// GetEnvironmentRegionByEnvAndRegion get
	GetEnvironmentRegionByEnvAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (*models.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by env
	ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error)
}

type dao struct{}

// NewDAO returns an instance of the default DAO
func NewDAO() DAO {
	return &dao{}
}

func (d *dao) CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(environment)

	return environment, result.Error
}

func (d *dao) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environments []*models.Environment

	result := db.Raw(common.EnvironmentListAll).Scan(&environments)

	return environments, result.Error
}

func (d *dao) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	result := db.Create(er)
	return er, result.Error
}

func (d *dao) ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var regions []string
	result := db.Raw(common.EnvironmentListRegion, env).Scan(&regions)

	return regions, result.Error
}

func (d *dao) GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error) {
	db, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var environmentRegion models.EnvironmentRegion
	result := db.Raw(common.EnvironmentRegionGetByID, id).First(&environmentRegion)

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

	return &environmentRegion, result.Error
}
