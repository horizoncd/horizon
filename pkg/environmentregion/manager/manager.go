package manager

import (
	"context"

	"github.com/horizoncd/horizon/pkg/environmentregion/dao"
	"github.com/horizoncd/horizon/pkg/environmentregion/models"
	regiondao "github.com/horizoncd/horizon/pkg/region/dao"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"gorm.io/gorm"
)

func New(db *gorm.DB) Manager {
	return &manager{
		envRegionDAO: dao.NewDAO(db),
		regionDAO:    regiondao.NewDAO(db),
	}
}

type Manager interface {
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *models.EnvironmentRegion) (
		*models.EnvironmentRegion, error)
	// ListByEnvironment list regions by env
	ListByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error)
	ListEnabledRegionsByEnvironment(ctx context.Context, env string) (regionmodels.RegionParts, error)
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error)
	GetByEnvironmentAndRegion(ctx context.Context, env, region string) (*models.EnvironmentRegion, error)
	GetDefaultRegionByEnvironment(ctx context.Context, env string) (*models.EnvironmentRegion, error)
	SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error
	// ListAllEnvironmentRegions list all environmentRegions
	ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error)
	DeleteByID(ctx context.Context, id uint) error
}

type manager struct {
	envRegionDAO dao.DAO
	regionDAO    regiondao.DAO
}

// DeleteByID implements Manager.
func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.DeleteByID(ctx, id)
}

func (m *manager) GetDefaultRegionByEnvironment(ctx context.Context, env string) (
	*models.EnvironmentRegion, error,
) {
	return m.envRegionDAO.GetDefaultRegionByEnvironment(ctx, env)
}

func (m *manager) CreateEnvironmentRegion(ctx context.Context,
	er *models.EnvironmentRegion,
) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.CreateEnvironmentRegion(ctx, er)
}

func (m *manager) ListByEnvironment(ctx context.Context, env string) ([]*models.EnvironmentRegion, error) {
	regions, err := m.envRegionDAO.ListRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func (m *manager) ListEnabledRegionsByEnvironment(ctx context.Context, env string) (
	regionmodels.RegionParts, error,
) {
	return m.envRegionDAO.ListEnabledRegionsByEnvironment(ctx, env)
}

func (m *manager) GetEnvironmentRegionByID(ctx context.Context, id uint) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByID(ctx, id)
}

func (m *manager) GetByEnvironmentAndRegion(ctx context.Context,
	env, region string,
) (*models.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByEnvAndRegion(ctx, env, region)
}

func (m *manager) SetEnvironmentRegionToDefaultByID(ctx context.Context, id uint) error {
	return m.envRegionDAO.SetEnvironmentRegionToDefaultByID(ctx, id)
}

func (m *manager) ListAllEnvironmentRegions(ctx context.Context) ([]*models.EnvironmentRegion, error) {
	return m.envRegionDAO.ListAllEnvironmentRegions(ctx)
}
