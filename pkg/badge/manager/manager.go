package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/pkg/badge/dao"
	"github.com/horizoncd/horizon/pkg/badge/models"
)

type Manager interface {
	Create(ctx context.Context, badge *models.Badge) (*models.Badge, error)
	Update(ctx context.Context, badge *models.Badge) (*models.Badge, error)

	UpdateByName(ctx context.Context, resourceType string,
		resourceID uint, name string, badge *models.Badge) (*models.Badge, error)
	List(ctx context.Context, resourceType string, resourceID uint) ([]*models.Badge, error)
	Get(ctx context.Context, id uint) (*models.Badge, error)
	GetByName(ctx context.Context, resourceType string, resourceID uint, name string) (*models.Badge, error)
	Delete(ctx context.Context, id uint) error
	DeleteByName(ctx context.Context, resourceType string, resourceID uint, name string) error
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

func (m *manager) Create(ctx context.Context, badge *models.Badge) (*models.Badge, error) {
	return m.dao.Create(ctx, badge)
}

func (m *manager) Update(ctx context.Context, badge *models.Badge) (*models.Badge, error) {
	return m.dao.Update(ctx, badge)
}

func (m *manager) UpdateByName(ctx context.Context, resourceType string,
	resourceID uint, name string, badge *models.Badge) (*models.Badge, error) {
	return m.dao.UpdateByName(ctx, resourceType, resourceID, name, badge)
}

func (m *manager) List(ctx context.Context, resourceType string, resourceID uint) ([]*models.Badge, error) {
	return m.dao.List(ctx, resourceType, resourceID)
}

func (m *manager) Get(ctx context.Context, id uint) (*models.Badge, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) GetByName(ctx context.Context, resourceType string,
	resourceID uint, name string) (*models.Badge, error) {
	return m.dao.GetByName(ctx, resourceType, resourceID, name)
}

func (m *manager) Delete(ctx context.Context, id uint) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) DeleteByName(ctx context.Context, resourceType string, resourceID uint, name string) error {
	return m.dao.DeleteByName(ctx, resourceType, resourceID, name)
}
