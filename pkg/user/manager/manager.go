package manager

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
	// GetByOIDCMeta get user oidcType and email
	GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error)
	// SearchUser search user by filter
	SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetUserByIDs(ctx context.Context, userIDs []uint) ([]models.User, error)
	GetUserMapByIDs(ctx context.Context, userIDs []uint) (map[uint]*models.User, error)
}

type manager struct {
	dao dao.DAO
}

func New() Manager {
	return &manager{dao: dao.NewDAO()}
}

func (m *manager) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return m.dao.Create(ctx, user)
}

func (m *manager) GetByOIDCMeta(ctx context.Context, oidcType, email string) (*models.User, error) {
	return m.dao.GetByOIDCMeta(ctx, oidcType, email)
}

func (m *manager) SearchUser(ctx context.Context, filter string, query *q.Query) (int, []models.User, error) {
	return m.dao.SearchUser(ctx, filter, query)
}

func (m *manager) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.dao.GetByEmail(ctx, email)
}

func (m *manager) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	users, err := m.dao.GetByIDs(ctx, []uint{userID})
	if err != nil {
		return nil, err
	}
	if users == nil || len(users) < 1 {
		return nil, nil
	}
	return &users[0], nil
}

func (m *manager) GetUserByIDs(ctx context.Context, userIDs []uint) ([]models.User, error) {
	return m.dao.GetByIDs(ctx, userIDs)
}

func (m *manager) GetUserMapByIDs(ctx context.Context, userIDs []uint) (map[uint]*models.User, error) {
	users, err := m.GetUserByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint]*models.User)
	for _, user := range users {
		userMap[user.ID] = &user
	}
	return userMap, nil
}
