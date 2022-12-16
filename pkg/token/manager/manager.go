package manager

import (
	"context"

	"github.com/horizoncd/horizon/pkg/token/dao"
	"github.com/horizoncd/horizon/pkg/token/models"
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
	return &manager{dao: dao.NewDAO(db)}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) CreateToken(ctx context.Context, token *models.Token) (*models.Token, error) {
	return m.dao.Create(ctx, token)
}

func (m *manager) LoadTokenByID(ctx context.Context, id uint) (*models.Token, error) {
	return m.dao.GetByID(ctx, id)
}

func (m *manager) LoadTokenByCode(ctx context.Context, code string) (*models.Token, error) {
	return m.dao.GetByCode(ctx, code)
}

func (m *manager) RevokeTokenByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}

func (m *manager) RevokeTokenByClientID(ctx context.Context, clientID string) error {
	return m.dao.DeleteByClientID(ctx, clientID)
}
