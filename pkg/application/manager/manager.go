package manager

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/lib/q"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/sets"

	"gorm.io/gorm"
)

var (
	// Mgr is the global application manager
	Mgr = New()
)

const _errCodeApplicationNotFound = errors.ErrorCode("ApplicationNotFound")
const _errCodeUserNotFound = errors.ErrorCode("UserNotFound")

type Manager interface {
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Application, error)
	GetByName(ctx context.Context, name string) (*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	// GetByNameFuzzilyByPagination get applications that fuzzily matching the given name
	GetByNameFuzzilyByPagination(ctx context.Context, name string, query q.Query) (int, []*models.Application, error)
	Create(ctx context.Context, application *models.Application, extraOwners []string) (*models.Application, error)
	UpdateByID(ctx context.Context, id uint, application *models.Application) (*models.Application, error)
	DeleteByID(ctx context.Context, id uint) error
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
	const op = "application manager: create application"
	users, err := m.userDAO.ListByEmail(ctx, extraOwners)
	if err != nil {
		return nil, err
	}

	if len(users) < len(extraOwners) {
		userSet := sets.NewString()
		for _, user := range users {
			userSet.Insert(user.Email)
		}
		for _, owner := range extraOwners {
			if !userSet.Has(owner) {
				return nil, errors.E(op, http.StatusNotFound, _errCodeUserNotFound,
					fmt.Sprintf("user with email %s not found, please login in horizon first.", owner))
			}
		}
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
