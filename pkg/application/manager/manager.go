package manager

import (
	"context"
	"net/http"

	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"gorm.io/gorm"
)

var (
	// Mgr is the global application manager
	Mgr = New()
)

const _errCodeApplicationNotFound = errors.ErrorCode("ApplicationNotFound")

type Manager interface {
	GetByName(ctx context.Context, name string) (*models.Application, error)
	// GetByNameFuzzily get applications that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error)
	Create(ctx context.Context, application *models.Application) (*models.Application, error)
	UpdateByName(ctx context.Context, name string, application *models.Application) (*models.Application, error)
	DeleteByName(ctx context.Context, name string) error
}

func New() Manager {
	return &manager{
		applicationDAO: applicationdao.NewDAO(),
		groupDAO:       groupdao.NewDAO(),
	}
}

type manager struct {
	applicationDAO applicationdao.DAO
	groupDAO       groupdao.DAO
}

func (m *manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Application, error) {
	return m.applicationDAO.GetByNameFuzzily(ctx, name)
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

func (m *manager) Create(ctx context.Context, application *models.Application) (*models.Application, error) {
	return m.applicationDAO.Create(ctx, application)
}

func (m *manager) UpdateByName(ctx context.Context, name string,
	application *models.Application) (*models.Application, error) {
	return m.applicationDAO.UpdateByName(ctx, name, application)
}

func (m *manager) DeleteByName(ctx context.Context, name string) error {
	return m.applicationDAO.DeleteByName(ctx, name)
}
