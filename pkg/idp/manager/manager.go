package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/idp/dao"
	"g.hz.netease.com/horizon/pkg/idp/models"
	"gorm.io/gorm"
)

type Manager interface {
	ListIDP(ctx context.Context) ([]*models.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error)
}

type manager struct {
	dao dao.DAO
}

func NewManager(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) ListIDP(ctx context.Context) ([]*models.IdentityProvider, error) {
	return m.dao.ListIDP(ctx)
}

func (m *manager) GetProviderByName(ctx context.Context, name string) (*models.IdentityProvider, error) {
	return m.dao.GetProviderByName(ctx, name)
}
