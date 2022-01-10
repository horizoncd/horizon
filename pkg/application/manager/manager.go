package manager

import (
	"context"
	"net/http"

	"g.hz.netease.com/horizon/lib/q"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

var (
	// Mgr is the global application manager
	Mgr = New()
)

const _errCodeApplicationNotFound = errors.ErrorCode("ApplicationNotFound")

type Manager interface {
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	// GetByNameFuzzilyByPagination get applications that fuzzily matching the given name
	GetByNameFuzzilyByPagination(ctx context.Context, name string, query q.Query) (int, []*models.Application, error)
	Create(ctx context.Context, application *models.Application, extraOwners []string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
	// ListUserAuthorizedByNameFuzzily list application which is authorized to the specified user.
	// 1. name is the application's fuzzily name.
	// 2. groupIDs is the groups' id which are authorized to the specified user.
	// 3. userInfo is the user id
	ListUserAuthorizedByNameFuzzily(ctx context.Context,
		name string, groupIDs []uint, userInfo uint, query *q.Query) (int, []*models.Application, error)
}

func New() Manager {
	return &manager{
		applicationDAO: applicationdao.NewDAO(),
		groupDAO:       groupdao.NewDAO(),
		userDAO:        userdao.NewDAO(),
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

func (m *manager) GetByNameFuzzilyByPagination(ctx context.Context, name string, query q.Query) (int,
	[]*models.Application, error) {
	return m.applicationDAO.GetByNameFuzzilyByPagination(ctx, name, query)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Application, error) {
	const op = "application manager: get by id"
	application, err := m.applicationDAO.GetByID(ctx, id)
	// TODO(gjq) error handing outside
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, _errCodeApplicationNotFound)
		}
		return nil, errors.E(op, err)
	}
	return application, nil
}

func (m *manager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error) {
	const op = "application manager: get by ids"
	applications, err := m.applicationDAO.GetByIDs(ctx, ids)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return applications, nil
}

func (m *manager) GetByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Application, error) {
	return m.applicationDAO.GetByGroupIDs(ctx, groupIDs)
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Application, error) {
	const op = "application manager: get by name"
	application, err := m.applicationDAO.GetByName(ctx, name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.E(op, http.StatusNotFound, _errCodeApplicationNotFound)
		}
		return nil, errors.E(op, err)
	}
	return application, nil
}

func (m *manager) Create(ctx context.Context, application *models.Application,
	extraOwners []string) (*models.Application, error) {
	users, err := m.userDAO.ListByEmail(ctx, extraOwners)
	if err != nil {
		return nil, err
	}

	return m.applicationDAO.Create(ctx, application, users)
}

func (m *manager) UpdateByID(ctx context.Context,
	id uint, application *models.Application) (*models.Application, error) {
	return m.applicationDAO.UpdateByID(ctx, id, application)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.applicationDAO.DeleteByID(ctx, id)
}

func (m *manager) ListUserAuthorizedByNameFuzzily(ctx context.Context,
	name string, groupIDs []uint, userInfo uint, query *q.Query) (int, []*models.Application, error) {
	return m.applicationDAO.ListUserAuthorizedByNameFuzzily(ctx, name, groupIDs, userInfo, query)
}
