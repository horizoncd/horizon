package user

import (
	"context"

	"g.hz.netease.com/horizon/pkg/lib/q"
)

var (
	// Mgr is the global user manager
	Mgr = New()
)

type Manager interface {
	// Create user
	Create(ctx context.Context, user *User) (*User, error)
	// GetByOIDCMeta get user by oidcID and oidcType
	GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*User, error)
	// SearchUser search user by filter
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []User, error)
}

type manager struct {
	dao DAO
}

func New() Manager {
	return &manager{dao: newDAO()}
}

func (m *manager) Create(ctx context.Context, user *User) (*User, error) {
	return m.dao.Create(ctx, user)
}

func (m *manager) GetByOIDCMeta(ctx context.Context, oidcID, oidcType string) (*User, error) {
	return m.dao.GetByOIDCMeta(ctx, oidcID, oidcType)
}

func (m *manager) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []User, error) {
	return m.dao.SearchUser(ctx, filter, query)
}
