package store

import (
	"github.com/horizoncd/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
)

type TokenStore interface {
	Create(ctx context.Context, token *models.Token) (*models.Token, error)
	DeleteByCode(ctx context.Context, code string) error
	DeleteByClientID(ctx context.Context, code string) error
	Get(ctx context.Context, code string) (*models.Token, error)
	DeleteByID(ctx context.Context, id uint) error
}

type OauthAppStore interface {
	CreateApp(ctx context.Context, client models.OauthApp) error
	GetApp(ctx context.Context, clientID string) (*models.OauthApp, error)
	DeleteApp(ctx context.Context, clientID string) error
	ListApp(ctx context.Context, ownerType models.OwnerType, ownerID uint) ([]models.OauthApp, error)
	UpdateApp(ctx context.Context, clientID string, app models.OauthApp) (*models.OauthApp, error)
	CreateSecret(ctx context.Context, secret *models.OauthClientSecret) (*models.OauthClientSecret, error)
	DeleteSecret(ctx context.Context, clientID string, clientSecretID uint) error
	DeleteSecretByClientID(ctx context.Context, clientID string) error
	ListSecret(ctx context.Context, clientID string) ([]models.OauthClientSecret, error)
}
