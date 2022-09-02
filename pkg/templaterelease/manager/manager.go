package manager

import (
	"context"

	amodels "g.hz.netease.com/horizon/pkg/application/models"
	cmodel "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/templaterelease/dao"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"gorm.io/gorm"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/templaterelease/manager/manager_mock.go -package=mock_manager
type Manager interface {
	// Create template release
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	// ListByTemplateName list all releases by template name
	ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error)
	// ListByTemplateID list all releases by template ID
	ListByTemplateID(ctx context.Context, id uint) ([]*models.TemplateRelease, error)
	// GetByTemplateNameAndRelease get release by template name and release name
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
	// GetByID gets release by releaseID
	GetByID(ctx context.Context, releaseID uint) (*models.TemplateRelease, error)
	GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error)
	GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error)
	UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error
	DeleteByID(ctx context.Context, id uint) error
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context,
	templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	return m.dao.Create(ctx, templateRelease)
}

func (m *manager) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	return m.dao.ListByTemplateName(ctx, templateName)
}
func (m *manager) ListByTemplateID(ctx context.Context, id uint) ([]*models.TemplateRelease, error) {
	return m.dao.ListByTemplateID(ctx, id)
}

func (m *manager) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (_ *models.TemplateRelease, err error) {
	const op = "template release manager: get by template name and release"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := m.dao.GetByTemplateNameAndRelease(ctx, templateName, release)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (m *manager) GetByID(ctx context.Context,
	releaseID uint) (*models.TemplateRelease, error) {
	return m.dao.GetByID(ctx, releaseID)
}

func (m *manager) GetRefOfApplication(ctx context.Context, id uint) ([]*amodels.Application, uint, error) {
	return m.dao.GetRefOfApplication(ctx, id)
}
func (m *manager) GetRefOfCluster(ctx context.Context, id uint) ([]*cmodel.Cluster, uint, error) {
	return m.dao.GetRefOfCluster(ctx, id)
}

func (m *manager) UpdateByID(ctx context.Context, releaseID uint, release *models.TemplateRelease) error {
	return m.dao.UpdateByID(ctx, releaseID, release)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.dao.DeleteByID(ctx, id)
}
