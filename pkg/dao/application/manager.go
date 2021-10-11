package application

import (
	"context"
)

var (
	// Mgr is the global application manager
	Mgr = New()
)

type Manager interface {
	GetByName(ctx context.Context, name string) (*Application, error)
	Create(ctx context.Context, application *Application) (*Application, error)
	UpdateByName(ctx context.Context, name string, application *Application) (*Application, error)
	DeleteByName(ctx context.Context, name string) error
}

func New() Manager {
	return &manager{dao: newDAO()}
}

type manager struct {
	dao DAO
}

func (m *manager) GetByName(ctx context.Context, name string) (*Application, error) {
	return m.dao.GetByName(ctx, name)
}

func (m *manager) Create(ctx context.Context, application *Application) (*Application, error) {
	return m.dao.Create(ctx, application)
}

func (m *manager) UpdateByName(ctx context.Context, name string,
	application *Application) (*Application, error) {
	return m.dao.UpdateByName(ctx, name, application)
}

func (m *manager) DeleteByName(ctx context.Context, name string) error {
	return m.dao.DeleteByName(ctx, name)
}
