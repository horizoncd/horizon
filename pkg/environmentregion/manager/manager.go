package manager

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	envdao "g.hz.netease.com/horizon/pkg/environment/dao"
	"g.hz.netease.com/horizon/pkg/environmentregion/dao"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

var (
	// Mgr is the global environment manager
	Mgr = New()
)

func New() Manager {
	return &manager{
		envDAO:       envdao.NewDAO(),
		envRegionDAO: dao.NewDAO(),
		regionDAO:    regiondao.NewDAO(),
	}
}

type Manager interface {
	// CreateEnvironmentRegion create a environmentRegion
	CreateEnvironmentRegion(ctx context.Context, er *envregionmodels.EnvironmentRegion) (
		*envregionmodels.EnvironmentRegion, error)
	// ListRegionsByEnvironment list regions by env
	ListRegionsByEnvironment(ctx context.Context, env string) ([]*regionmodels.Region, error)
	GetEnvironmentRegionByID(ctx context.Context, id uint) (*envregionmodels.EnvironmentRegion, error)
	GetByEnvironmentAndRegion(ctx context.Context, env, region string) (*envregionmodels.EnvironmentRegion, error)
	// ListAllEnvironmentRegions list all environmentRegions
	ListAllEnvironmentRegions(ctx context.Context) ([]*envregionmodels.EnvironmentRegion, error)
	// UpdateEnvironmentRegionByID update environmentRegion by id
	UpdateEnvironmentRegionByID(ctx context.Context, id uint, environmentRegion *envregionmodels.EnvironmentRegion) error
}

type manager struct {
	envDAO       envdao.DAO
	envRegionDAO dao.DAO
	regionDAO    regiondao.DAO
}

func (m *manager) UpdateEnvironmentRegionByID(ctx context.Context, id uint,
	environmentRegion *envregionmodels.EnvironmentRegion) error {
	return m.envRegionDAO.UpdateEnvironmentRegionByID(ctx, id, environmentRegion)
}

func (m *manager) ListAllEnvironmentRegions(ctx context.Context) ([]*envregionmodels.EnvironmentRegion, error) {
	return m.envRegionDAO.ListAllEnvironmentRegions(ctx)
}

func (m *manager) CreateEnvironmentRegion(ctx context.Context,
	er *envregionmodels.EnvironmentRegion) (*envregionmodels.EnvironmentRegion, error) {
	_, err := m.envRegionDAO.GetEnvironmentRegionByEnvAndRegion(ctx,
		er.EnvironmentName, er.RegionName)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok || e.Source != herrors.EnvironmentRegionInDB {
			return nil, err
		}
	} else {
		return nil, perror.Wrap(herrors.ErrNameConflict, "already exists")
	}

	return m.envRegionDAO.CreateEnvironmentRegion(ctx, er)
}

func (m *manager) ListRegionsByEnvironment(ctx context.Context, env string) ([]*regionmodels.Region, error) {
	regionNames, err := m.envRegionDAO.ListRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}

	regions, err := m.regionDAO.ListByNames(ctx, regionNames)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func (m *manager) GetEnvironmentRegionByID(ctx context.Context, id uint) (*envregionmodels.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByID(ctx, id)
}

func (m *manager) GetByEnvironmentAndRegion(ctx context.Context,
	env, region string) (*envregionmodels.EnvironmentRegion, error) {
	return m.envRegionDAO.GetEnvironmentRegionByEnvAndRegion(ctx, env, region)
}
