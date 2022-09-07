package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/idp/dao"
	"g.hz.netease.com/horizon/pkg/idp/models"
	"gorm.io/gorm"
)

type Manager interface {
	List(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
	Create(ctx context.Context, idp *models.IdentityProvider) (*models.IdentityProvider, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error)
	Update(ctx context.Context, id uint, param *models.IdentityProvider) (*models.IdentityProvider, error)
}

type manager struct {
	dao dao.DAO
}

func NewManager(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) List(ctx context.Context) ([]*models.IdentityProvider, error) {
	return m.dao.List(ctx)
}

func (m *manager) GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error) {
	return m.dao.GetProviderByName(ctx, name)
}

func (m *manager) Create(ctx context.Context,
	idp *models.IdentityProvider) (*models.IdentityProvider, error) {
	return m.dao.Create(ctx, idp)
}

func (m *manager) Delete(ctx context.Context, id uint) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.IdentityProvider, error) {
	return m.dao.GetByID(ctx, id)
}

func (m *manager) Update(ctx context.Context,
	id uint, param *models.IdentityProvider) (*models.IdentityProvider, error) {
	return m.dao.Update(ctx, id, param)
}
