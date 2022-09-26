package manager

import (
	"context"

	amodels "g.hz.netease.com/horizon/pkg/application/models"
	cmodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/template/dao"
	"g.hz.netease.com/horizon/pkg/template/models"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/template/manager/manager_mock.go -package=mock_manager
type Manager interface {
	// Create template
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	// List all template
	List(ctx context.Context) ([]*models.Template, error)
	// ListByGroupID lists all template by group ID
	ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error)
	// DeleteByID deletes template by ID
	DeleteByID(ctx context.Context, id uint) error
	// GetByID gets a template by ID
	GetByID(ctx context.Context, id uint) (*models.Template, error)
	// GetByName gets a template by name
	GetByName(ctx context.Context, name string) (*models.Template, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*cmodels.Cluster, uint, error)
	UpdateByID(ctx context.Context, id uint, template *models.Template) error
	ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
	ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error)
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	return m.dao.Create(ctx, template)
}

func (m *manager) List(ctx context.Context) ([]*models.Template, error) {
	return m.dao.List(ctx)
}

func (m *manager) ListByGroupID(ctx context.Context, groupID uint) ([]*models.Template, error) {
	return m.dao.ListByGroupID(ctx, groupID)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}

func (m *manager) GetByID(ctx context.Context, id uint) (*models.Template, error) {
	return m.dao.GetByID(ctx, id)
}

func (m *manager) GetByName(ctx context.Context, name string) (*models.Template, error) {
	return m.dao.GetByName(ctx, name)
}

func (m *manager) GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error) {
	return m.dao.GetRefOfApplication(ctx, id)
}

func (m *manager) GetRefOfCluster(ctx context.Context, id uint) ([]*cmodels.Cluster, uint, error) {
	return m.dao.GetRefOfCluster(ctx, id)
}

func (m *manager) UpdateByID(ctx context.Context, id uint, template *models.Template) error {
	return m.dao.UpdateByID(ctx, id, template)
}

func (m *manager) ListByGroupIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	return m.dao.ListByGroupIDs(ctx, ids)
}

func (m *manager) ListByIDs(ctx context.Context, ids []uint) ([]*models.Template, error) {
	return m.dao.ListByIDs(ctx, ids)
}
