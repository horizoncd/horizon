package manager

import (
	"context"

	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
)

var (
	// Mgr is the global application manager
	Mgr = New()
)

type Manager interface {
	GetByName(ctx context.Context, name string) (*models.Application, error)
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

func (m *manager) GetByName(ctx context.Context, name string) (*models.Application, error) {
	return m.applicationDAO.GetByName(ctx, name)
}

func (m *manager) Create(ctx context.Context, application *models.Application) (*models.Application, error) {
	// TODO（gjq） 校验group
	return m.applicationDAO.Create(ctx, application)
}

func (m *manager) UpdateByName(ctx context.Context, name string,
	application *models.Application) (*models.Application, error) {
	return m.applicationDAO.UpdateByName(ctx, name, application)
}

func (m *manager) DeleteByName(ctx context.Context, name string) error {
	return m.applicationDAO.DeleteByName(ctx, name)
}
