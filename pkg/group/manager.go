package group

import (
	"context"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
)

var (
	// Mgr is the global group manager
	Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, group *models.Group) (uint, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*models.Group, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error)
	GetByTraversalIDs(ctx context.Context, traversalIDs string) ([]*models.Group, error)
	GetByPath(ctx context.Context, path string) (*models.Group, error)
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	GetByFullNamesRegexpFuzzily(ctx context.Context, names *[]string) ([]*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error)
	List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error)
}

type manager struct {
	dao dao.DAO
}

func (m manager) GetByIDsOrderByIDDesc(ctx context.Context, ids []int) ([]*models.Group, error) {
	return m.dao.GetByIDsOrderByIDDesc(ctx, ids)
}

func (m manager) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	return m.dao.GetByPaths(ctx, paths)
}

func (m manager) GetByIDs(ctx context.Context, ids []int) ([]*models.Group, error) {
	return m.dao.GetByIDs(ctx, ids)
}

// GetByTraversalIDs traversalIDs: 1,2,3
func (m manager) GetByTraversalIDs(ctx context.Context, traversalIDs string) ([]*models.Group, error) {
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]int, len(splitIds))
	for i, id := range splitIds {
		ids[i], _ = strconv.Atoi(id)
	}

	return m.GetByIDs(ctx, ids)
}

func (m manager) GetByFullNamesRegexpFuzzily(ctx context.Context, names *[]string) ([]*models.Group, error) {
	return m.dao.GetByFullNamesRegexpFuzzily(ctx, names)
}

func (m manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	return m.dao.GetByNameFuzzily(ctx, name)
}

func (m manager) Create(ctx context.Context, group *models.Group) (uint, error) {
	return m.dao.Create(ctx, group)
}

func (m manager) Delete(ctx context.Context, id uint) error {
	return m.dao.Delete(ctx, id)
}

func (m manager) Get(ctx context.Context, id uint) (*models.Group, error) {
	return m.dao.Get(ctx, id)
}

func (m manager) GetByPath(ctx context.Context, path string) (*models.Group, error) {
	return m.dao.GetByPath(ctx, path)
}

func (m manager) Update(ctx context.Context, group *models.Group) error {
	return m.dao.Update(ctx, group)
}

func (m manager) ListWithoutPage(ctx context.Context, query *q.Query) ([]*models.Group, error) {
	return m.dao.ListWithoutPage(ctx, query)
}

func (m manager) List(ctx context.Context, query *q.Query) ([]*models.Group, int64, error) {
	return m.dao.List(ctx, query)
}

func New() Manager {
	return &manager{dao: dao.New()}
}
