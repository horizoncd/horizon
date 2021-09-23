package application

import (
	"context"

	"g.hz.netease.com/horizon/pkg/application/dao"
	"g.hz.netease.com/horizon/pkg/application/models"
)

var (
	// Mgr is the global gitlab manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, application *models.Application) (*models.Application, error)
}

func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, application *models.Application) (*models.Application, error) {
	return m.dao.Create(ctx, application)
}
