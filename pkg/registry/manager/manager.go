package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/registry/dao"
	"g.hz.netease.com/horizon/pkg/registry/models"
	"gorm.io/gorm"
)

type Manager interface {
	// Create a harbor
	Create(ctx context.Context, harbor *models.Registry) (uint, error)
	// UpdateByID update a harbor
	UpdateByID(ctx context.Context, id uint, harbor *models.Registry) error
	// DeleteByID delete a harbor by id
	DeleteByID(ctx context.Context, id uint) error
	// GetByID get by id
	GetByID(ctx context.Context, id uint) (*models.Registry, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) ([]*models.Registry, error)
}

type manager struct {
	registryDAO dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		registryDAO: dao.NewDAO(db),
	}
}

func (m manager) Create(ctx context.Context, harbor *models.Registry) (uint, error) {
	return m.registryDAO.Create(ctx, harbor)
}

func (m manager) GetByID(ctx context.Context, id uint) (*models.Registry, error) {
	return m.registryDAO.GetByID(ctx, id)
}

func (m manager) ListAll(ctx context.Context) ([]*models.Registry, error) {
	return m.registryDAO.ListAll(ctx)
}

func (m manager) UpdateByID(ctx context.Context, id uint, harbor *models.Registry) error {
	return m.registryDAO.UpdateByID(ctx, id, harbor)
}

func (m manager) DeleteByID(ctx context.Context, id uint) error {
	return m.registryDAO.DeleteByID(ctx, id)
}
