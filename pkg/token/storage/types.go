package storage

import (
	"context"

	"github.com/horizoncd/horizon/pkg/token/models"
)

type Storage interface {
	Create(ctx context.Context, token *models.Token) (*models.Token, error)
	GetByID(ctx context.Context, id uint) (*models.Token, error)
	GetByCode(ctx context.Context, code string) (*models.Token, error)
	DeleteByID(ctx context.Context, id uint) error
	DeleteByCode(ctx context.Context, code string) error
	DeleteByClientID(ctx context.Context, clientID string) error
}
