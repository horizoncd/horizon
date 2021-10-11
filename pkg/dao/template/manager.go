package template

import (
	"context"
)

var (
	// Mgr is the global template manager
	Mgr = New()
)

type Manager interface {
	// Create template
	Create(ctx context.Context, template *Template) (*Template, error)
	// List all template
	List(ctx context.Context) ([]Template, error)
}

func New() Manager {
	return &manager{dao: newDAO()}
}

type manager struct {
	dao DAO
}

func (m *manager) Create(ctx context.Context, template *Template) (*Template, error) {
	return m.dao.Create(ctx, template)
}

func (m *manager) List(ctx context.Context) ([]Template, error) {
	return m.dao.List(ctx)
}
