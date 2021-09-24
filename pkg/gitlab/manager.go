package gitlab

import (
	"context"

	"g.hz.netease.com/horizon/pkg/gitlab/dao"
	"g.hz.netease.com/horizon/pkg/gitlab/models"
)

var (
	// Mgr is the global gitlab manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, gitlab *models.Gitlab) (*models.Gitlab, error)
	List(ctx context.Context) ([]models.Gitlab, error)
	GetByName(ctx context.Context, name string) (*models.Gitlab, error)
}

func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, gitlab *models.Gitlab) (*models.Gitlab, error) {
	return m.dao.Create(ctx, gitlab)
}

func (m *manager) List(ctx context.Context) ([]models.Gitlab, error) {
	return m.dao.List(ctx)
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Gitlab, error) {
	return m.dao.GetByName(ctx, name)
}
