package manager

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/pkg/environment/dao"
	"g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"gorm.io/gorm"
)

var (
	// Mgr is the global environment manager
	Mgr = New()
)

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type Manager interface {
	EnvironmentManager
	EnvironmentRegionManager
}

type EnvironmentManager interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*models.Environment, error)
}

type EnvironmentRegionManager interface {
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (*models.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by env
	ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error)
}

type manager struct {
	dao dao.DAO
}

func (m *manager) CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	return m.dao.CreateEnvironment(ctx, environment)
}

func (m *manager) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	return m.dao.ListAllEnvironment(ctx)
}

func (m *manager) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	const op = "environment manager: create environmentRegion"

	_, err := m.dao.GetEnvironmentRegionByEnvAndRegion(ctx,
		er.EnvironmentName, er.RegionName)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else {
		return nil, errors.E(op, http.StatusConflict, errors.ErrorCode("AlreadyExists"))
	}

	return m.dao.CreateEnvironmentRegion(ctx, er)
}

func (m *manager) ListRegionsByEnvironment(ctx context.Context, env string) ([]string, error) {
	return m.dao.ListRegionsByEnvironment(ctx, env)
}
