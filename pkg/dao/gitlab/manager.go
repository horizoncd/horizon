package gitlab

import (
	"context"
)

var (
	// Mgr is the global gitlab manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, gitlab *Gitlab) (*Gitlab, error)
	List(ctx context.Context) ([]Gitlab, error)
	GetByName(ctx context.Context, name string) (*Gitlab, error)
}

func New() Manager {
	return &manager{dao: newDAO()}
}

type manager struct {
	dao DAO
}

func (m *manager) Create(ctx context.Context, gitlab *Gitlab) (*Gitlab, error) {
	return m.dao.Create(ctx, gitlab)
}

func (m *manager) List(ctx context.Context) ([]Gitlab, error) {
	return m.dao.List(ctx)
}

func (m *manager) GetByName(ctx context.Context, name string) (*Gitlab, error) {
	return m.dao.GetByName(ctx, name)
}
