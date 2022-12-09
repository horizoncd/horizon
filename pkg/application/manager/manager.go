package manager

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/application/manager/manager.go -package=mock_manager
type Manager interface {
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	Create(ctx context.Context, application *models.Application,
		extraMembers map[string]string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
	Transfer(ctx context.Context, id uint, groupID uint) error
	List(ctx context.Context, groupIDs []uint, query *q.Query) (int, []*models.Application, error)
}

func New(db *gorm.DB) Manager {
	return &manager{
		applicationDAO: applicationdao.NewDAO(db),
		groupDAO:       groupdao.NewDAO(db),
		userDAO:        userdao.NewDAO(db),
	}
}

type manager struct {
	applicationDAO applicationdao.DAO
	groupDAO       groupdao.DAO
	userDAO        userdao.DAO
}

func (m *manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error) {
	return m.applicationDAO.GetByNameFuzzily(ctx, name)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Application, error) {
	application, err := m.applicationDAO.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *manager) GetByIDIncludeSoftDelete(ctx context.Context, id uint) (*models.Application, error) {
	application, err := m.applicationDAO.GetByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *manager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error) {
	applications, err := m.applicationDAO.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (m *manager) GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error) {
	return m.applicationDAO.GetByGroupIDs(ctx, groupIDs)
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Application, error) {
	application, err := m.applicationDAO.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (m *manager) Create(ctx context.Context, application *models.Application,
	extraMembers map[string]string) (*models.Application, error) {
	emails := make([]string, 0, len(extraMembers))
	for k := range extraMembers {
		emails = append(emails, k)
	}

	users, err := m.userDAO.ListByEmail(ctx, emails)
	if err != nil {
		return nil, err
	}

	extraMembersWithUser := make(map[*usermodels.User]string)
	for _, user := range users {
		extraMembersWithUser[user] = extraMembers[user.Email]
	}

	return m.applicationDAO.Create(ctx, application, extraMembersWithUser)
}

func (m *manager) UpdateByID(ctx context.Context,
	id uint, application *models.Application) (*models.Application, error) {
	return m.applicationDAO.UpdateByID(ctx, id, application)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.applicationDAO.DeleteByID(ctx, id)
}

func (m *manager) Transfer(ctx context.Context, id uint, groupID uint) error {
	return m.applicationDAO.TransferByID(ctx, id, groupID)
}

func (m *manager) List(ctx context.Context, groupIDs []uint, query *q.Query) (int, []*models.Application, error) {
	return m.applicationDAO.List(ctx, groupIDs, query)
}
