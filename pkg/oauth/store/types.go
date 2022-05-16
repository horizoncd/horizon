package store

import (
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
)

type TokenStore interface {
	Create(ctx context.Context, token *models.Token) (*models.Token, error)
	DeleteByCode(ctx context.Context, code string) (*models.Token, error)
	DeleteByClientID(ctx context.Context, code string) error
	Get(ctx context.Context, code string) (*models.Token, error)
}

type ClientAndSecretStore interface {
	CreateClient(ctx context.Context, client models.OauthServerInfo) error
	GetClient(ctx context.Context, clientID string) (models.OauthServerInfo, error)
	CreateSecret(ctx context.Context, secret *models.ClientSecret) error
	DeleteSecret(ctx context.Context, clientID string, clientSecretID uint) error
	ListSecret(ctx context.Context, clientID string) ([]models.ClientSecret, error)
}
