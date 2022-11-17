package manager

import (
	"context"

	"gorm.io/gorm"

	"g.hz.netease.com/horizon/lib/q"
	dao "g.hz.netease.com/horizon/pkg/accesstoken/dao"
	"g.hz.netease.com/horizon/pkg/accesstoken/models"
	oauthmodels "g.hz.netease.com/horizon/pkg/oauth/models"
	oauthstore "g.hz.netease.com/horizon/pkg/oauth/store"
)

type Manager interface {
	CreateAccessToken(context.Context, *oauthmodels.Token) (*oauthmodels.Token, error)
	DeleteAccessToken(context.Context, uint) error
	ListAccessTokensByResource(context.Context, string, uint, *q.Query) ([]*models.AccessToken, int, error)
	ListPersonalAccessTokens(context.Context, *q.Query) ([]*models.AccessToken, int, error)
	GetAccessToken(context.Context, uint) (*oauthmodels.Token, error)
}

type manager struct {
	tokenstore oauthstore.TokenStore
	dao        dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao:        dao.NewDAO(db),
		tokenstore: oauthstore.NewTokenStore(db),
	}
}

func (m *manager) CreateAccessToken(ctx context.Context, token *oauthmodels.Token) (*oauthmodels.Token, error) {
	return m.tokenstore.Create(ctx, token)
}

func (m *manager) DeleteAccessToken(ctx context.Context, id uint) error {
	return m.tokenstore.DeleteByID(ctx, id)
}

func (m *manager) ListAccessTokensByResource(ctx context.Context, resourceType string,
	resourceID uint, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListAccessTokensByResource(ctx, resourceType, resourceID, query)
}

func (m *manager) ListPersonalAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListPersonalAccessTokens(ctx, query)
}

func (m *manager) GetAccessToken(ctx context.Context, id uint) (*oauthmodels.Token, error) {
	return m.dao.GetAccessToken(ctx, id)
}
