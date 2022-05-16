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

type ClientInfo interface {
	GetID() string
	GetSecrets() []string
	GetAuthorizationCallbackURL() string
}

type ClientAndSecretStore interface {
	CreateClient(ctx context.Context, info ClientInfo) error
	GetClient(ctx context.Context, clientID string) (ClientInfo, error)
	CreateSecret(ctx context.Context, info *models.ClientSecret) error
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error
}
