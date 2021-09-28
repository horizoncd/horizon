package application

import (
	"context"

	"g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
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
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Application, error) {
	return m.dao.GetByName(ctx, name)
}

func (m *manager) Create(ctx context.Context, application *models.Application) (*models.Application, error) {
	return m.dao.Create(ctx, application)
}

func (m *manager) UpdateByName(ctx context.Context, name string,
	application *models.Application) (*models.Application, error) {
	return m.dao.UpdateByName(ctx, name, application)
}

func (m *manager) DeleteByName(ctx context.Context, name string) error {
	return m.dao.DeleteByName(ctx, name)
}
