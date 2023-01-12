package manager

import (
	"context"

	"github.com/horizoncd/horizon/pkg/token/models"
	"github.com/horizoncd/horizon/pkg/token/storage"
	"gorm.io/gorm"
)

type Manager interface {
	CreateToken(context.Context, *models.Token) (*models.Token, error)
	LoadTokenByID(context.Context, uint) (*models.Token, error)
	LoadTokenByCode(ctx context.Context, code string) (*models.Token, error)
	RevokeTokenByID(context.Context, uint) error
	RevokeTokenByClientID(ctx context.Context, clientID string) error
}

func New(db *gorm.DB) Manager {
	return &manager{storage: storage.NewStorage(db)}
}

type manager struct {
	storage storage.Storage
}

func (m *manager) CreateToken(ctx context.Context, token *models.Token) (*models.Token, error) {
	return m.storage.Create(ctx, token)
}

func (m *manager) LoadTokenByID(ctx context.Context, id uint) (*models.Token, error) {
	return m.storage.GetByID(ctx, id)
}

func (m *manager) LoadTokenByCode(ctx context.Context, code string) (*models.Token, error) {
	return m.storage.GetByCode(ctx, code)
}

func (m *manager) RevokeTokenByID(ctx context.Context, id uint) error {
	return m.storage.DeleteByID(ctx, id)
}

func (m *manager) RevokeTokenByClientID(ctx context.Context, clientID string) error {
	return m.storage.DeleteByClientID(ctx, clientID)
}
