package user

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/user/models"
)

var (
	// Mgr is the global user manager
	Mgr = New()
)

type Manager interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	// GetByOIDCMeta get user by oidcID and oidcType
	GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*models.User, error)
	// SearchUser search user by filter
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error)
}

type manager struct {
	dao dao.DAO
}

func New() Manager {
	return &manager{dao: dao.New()}
}

func (m *manager) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return m.dao.Create(ctx, user)
}

func (m *manager) GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*models.User, error) {
	return m.dao.GetByOIDCMeta(ctx, oidcID, oidcType)
}

func (m *manager) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error) {
	return m.dao.SearchUser(ctx, filter, query)
}
