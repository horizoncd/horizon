package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/environment/dao"
	"g.hz.netease.com/horizon/pkg/environment/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
)

var (
	// Mgr is the global environment manager
	Mgr = New()
)

func New() Manager {
	return &manager{
		envDAO:    dao.NewDAO(),
		regionDAO: regiondao.NewDAO(),
	}
}

type Manager interface {
	// CreateEnvironment create a environment
	CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	// ListAllEnvironment list all environments
	ListAllEnvironment(ctx context.Context) ([]*models.Environment, error)
	// UpdateByID update environment by id
	UpdateByID(ctx context.Context, id uint, environment *models.Environment) error
	// DeleteByID delete environment by id
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Environment, error)
}

type manager struct {
	envDAO    dao.DAO
	regionDAO regiondao.DAO
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Environment, error) {
	return m.envDAO.GetByID(ctx, id)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.envDAO.DeleteByID(ctx, id)
}

func (m *manager) UpdateByID(ctx context.Context, id uint, environment *models.Environment) error {
	return m.envDAO.UpdateByID(ctx, id, environment)
}

func (m *manager) CreateEnvironment(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	return m.envDAO.CreateEnvironment(ctx, environment)
}

func (m *manager) ListAllEnvironment(ctx context.Context) ([]*models.Environment, error) {
	return m.envDAO.ListAllEnvironment(ctx)
}
