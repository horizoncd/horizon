package manager

import (
	"context"

	"g.hz.netease.com/horizon/pkg/template/dao"
	"g.hz.netease.com/horizon/pkg/template/models"
)

var (
	// Mgr is the global template manager
	Mgr = New()
)

type Manager interface {
	// Create template
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	// List all template
	List(ctx context.Context) ([]models.Template, error)
}

func New() Manager {
	return &manager{dao: dao.NewDAO()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	return m.dao.Create(ctx, template)
}

func (m *manager) List(ctx context.Context) ([]models.Template, error) {
	return m.dao.List(ctx)
}
