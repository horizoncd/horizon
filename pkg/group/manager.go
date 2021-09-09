package group

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
)

var (
	// Mgr is the global group manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, group *models.Group) (int64, error)
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*models.Group, error)
	GetByPath(ctx context.Context, path string) (*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	List(ctx context.Context, query *q.Query) ([]*models.Group, error)
}

type manager struct {
	dao dao.DAO
}

func (m manager) Create(ctx context.Context, group *models.Group) (int64, error) {
	return m.dao.Create(ctx, group)
}

func (m manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

func (m manager) Get(ctx context.Context, id int64) (*models.Group, error) {
	return m.dao.Get(ctx, id)
}

func (m manager) GetByPath(ctx context.Context, path string) (*models.Group, error) {
	return m.dao.GetByPath(ctx, path)
}

func (m manager) Update(ctx context.Context, group *models.Group) error {
	return m.dao.Update(ctx, group)
}

func (m manager) List(ctx context.Context, query *q.Query) ([]*models.Group, error) {
	return m.dao.List(ctx, query)
}

func New() Manager {
	return &manager{dao: dao.New()}
}
