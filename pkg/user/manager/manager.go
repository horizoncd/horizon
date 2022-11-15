package manager

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/user/manager/manager_mock.go -package=mock_manager
type Manager interface {
	// Create user
	Create(ctx context.Context, user *models.User) (*models.User, error)
	List(ctx context.Context, query *q.Query) (int64, []*models.User, error)
	GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetUserByIDs(ctx context.Context, userIDs []uint) ([]models.User, error)
	GetUserMapByIDs(ctx context.Context, userIDs []uint) (map[uint]*models.User, error)
	ListByEmail(ctx context.Context, emails []string) ([]*models.User, error)
	UpdateByID(ctx context.Context, id uint, db *models.User) (*models.User, error)
}

type manager struct {
	dao dao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

func (m *manager) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return m.dao.Create(ctx, user)
}

func (m *manager) List(ctx context.Context, query *q.Query) (int64, []*models.User, error) {
	return m.dao.List(ctx, query)
}

func (m *manager) ListByEmail(ctx context.Context, emails []string) ([]*models.User, error) {
	return m.dao.ListByEmail(ctx, emails)
}

func (m *manager) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	return m.dao.GetByID(ctx, userID)
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
		tmp := user
		userMap[user.ID] = &tmp
	}
	return userMap, nil
}

func (m *manager) GetUserByIDP(ctx context.Context, email string, idp string) (*models.User, error) {
	return m.dao.GetUserByIDP(ctx, email, idp)
}

func (m *manager) UpdateByID(ctx context.Context, id uint, db *models.User) (*models.User, error) {
	return m.dao.UpdateByID(ctx, id, db)
}
