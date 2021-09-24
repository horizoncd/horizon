package templaterelease

import (
	"context"

	"g.hz.netease.com/horizon/pkg/templaterelease/dao"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
)

var (
	// Mgr is the global template release manager
	Mgr = New()
)

type Manager interface {
	// Create template release
	Create(ctx context.Context, user *models.TemplateRelease) (*models.TemplateRelease, error)
	// ListByTemplateName list all releases by template name
	ListByTemplateName(ctx context.Context, templateName string) ([]models.TemplateRelease, error)
	// GetByTemplateNameAndRelease get release by template name and release name
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
}

func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, template *models.TemplateRelease) (*models.TemplateRelease, error) {
	return m.dao.Create(ctx, template)
}

func (m *manager) ListByTemplateName(ctx context.Context, templateName string) ([]models.TemplateRelease, error) {
	return m.dao.ListByTemplateName(ctx, templateName)
}

func (m *manager) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (*models.TemplateRelease, error) {
	return m.dao.GetByTemplateNameAndRelease(ctx, templateName, release)
}
