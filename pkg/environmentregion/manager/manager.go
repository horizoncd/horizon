package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environmentregion/dao"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
)

var (
	// Mgr is the global environment manager
	Mgr = New()
)

func New() Manager {
	return &manager{
		envRegionDAO: dao.NewDAO(),
		regionDAO:    regiondao.NewDAO(),
	}
}

type Manager interface {
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (
		*models.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by env
	ListRegionsByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error)
	ListEnabledRegionsByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error)
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	GetByEnvironmentAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	GetDefaultRegionByEnvironment(ctx context.Context, env string) (*models.EnvironmentRegion, error)
	GetDefaultRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	EnableEnvironmentRegionByID(ctx context.Context, id uint) error
	DisableEnvironmentRegionByID(ctx context.Context, id uint) error
	SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error
	// ListAllEnvironmentRegions list all environmentRegions
	ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
}

type manager struct {
	envRegionDAO dao.DAO
	regionDAO    regiondao.DAO
}

func (m *manager) GetDefaultRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetDefaultRegions(ctx)
}

func (m *manager) GetDefaultRegionByEnvironment(ctx context.Context, env string) (
	*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetDefaultRegionByEnvironment(ctx, env)
}

func (m *manager) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.CreateEnvironmentRegion(ctx, er)
}

func (m *manager) ListRegionsByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error) {
	regions, err := m.envRegionDAO.ListRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func (m *manager) ListEnabledRegionsByEnvironment(ctx context.Context, env string) (
	[]*models.EnvironmentRegion, error) {
	regions, err := m.envRegionDAO.ListEnabledRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func (m *manager) GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByID(ctx, id)
}

func (m *manager) GetByEnvironmentAndRegion(ctx context.Context,
	env, region string) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByEnvAndRegion(ctx, env, region)
}

func (m *manager) EnableEnvironmentRegionByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.EnableEnvironmentRegionByID(ctx, id)
}

func (m *manager) DisableEnvironmentRegionByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.DisableEnvironmentRegionByID(ctx, id)
}

func (m *manager) SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.SetEnvironmentRegionToDefaultByID(ctx, id)
}

func (m *manager) ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	return m.envRegionDAO.ListAllEnvironmentRegions(ctx)
}
