package manager

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/oauth/generate"
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	"golang.org/x/net/context"
)

type AuthorizeGenerateRequest struct {
	ClientID    string
	RedirectURL string
	State       string

	Request *http.Request
}

type AccessTokenGenerateRequest struct {
	ClientID     string
	ClientSecret string
	Code         string // authorization code
	RedirectURL  string

	Request *http.Request
}

type Manager interface {
	CreateOauthApp(ctx context.Context, info models.OauthServerInfo) error
	GetOAuthApp(ctx context.Context, clientID string) (models.OauthServerInfo, error)
	DeleteOauthApp(ctx context.Context, clientID string) error

	CreateSecret(ctx context.Context, info models.ClientSecret) error
	DeleteSecret(ctx context.Context, ClientID string, clientSecretID uint) error

	GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (models.Token, error)
	GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest) (models.Token, error)
	RevokeAllAccessToken(ctx context.Context, clientID string) error
	LoadAccessToken(ctx context.Context, AccessToken string) (models.Token, error)
}

type manager struct {
	tokenStore            store.TokenStore
	clientAndSecretStore  store.ClientAndSecretStore
	accessGenerate        generate.AccessTokenCodeGenerate
	authorizationGenerate generate.AuthorizationCodeGenerate
	// the code expire time and token expire time
}

func (m *manager) GenAuthorizeCode(ctx context.Context, req *AuthorizeGenerateRequest) (models.Token, error) {
	return models.Token{}, nil
}
func (m *manager) GenAccessToken(ctx context.Context, req *AccessTokenGenerateRequest) (models.Token, error) {
	return models.Token{}, nil
}
func (m *manager) RevokeAllAccessToken(ctx context.Context, clientID string) error {
	return nil
}
func (m *manager) LoadAccessToken(ctx context.Context, AccessToken string) (models.Token, error) {
	return models.Token{}, nil
}
