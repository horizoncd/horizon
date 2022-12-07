package manager

import (
	"context"

	"gorm.io/gorm"

	"github.com/horizoncd/horizon/lib/q"
	dao "github.com/horizoncd/horizon/pkg/accesstoken/dao"
	"github.com/horizoncd/horizon/pkg/accesstoken/models"
	oauthmodels "github.com/horizoncd/horizon/pkg/oauth/models"
	oauthstore "github.com/horizoncd/horizon/pkg/oauth/store"
)

type Manager interface {
	ListAccessTokensByResource(context.Context, string, uint, *q.Query) ([]*models.AccessToken, int, error)
	ListPersonalAccessTokens(context.Context, *q.Query) ([]*models.AccessToken, int, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

func (m *manager) ListAccessTokensByResource(ctx context.Context, resourceType string,
	resourceID uint, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListAccessTokensByResource(ctx, resourceType, resourceID, query)
}

func (m *manager) ListPersonalAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error) {
	return m.dao.ListPersonalAccessTokens(ctx, query)
}
