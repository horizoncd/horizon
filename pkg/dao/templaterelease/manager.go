package templaterelease

import (
	"context"
)

var (
	// Mgr is the global template release manager
	Mgr = New()
)

type Manager interface {
	// Create template release
	Create(ctx context.Context, templateRelease *TemplateRelease) (*TemplateRelease, error)
	// ListByTemplateName list all releases by template name
	ListByTemplateName(ctx context.Context, templateName string) ([]TemplateRelease, error)
	// GetByTemplateNameAndRelease get release by template name and release name
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*TemplateRelease, error)
}

func New() Manager {
	return &manager{dao: newDAO()}
}

type manager struct {
	dao DAO
}

func (m *manager) Create(ctx context.Context, templateRelease *TemplateRelease) (*TemplateRelease, error) {
	return m.dao.Create(ctx, templateRelease)
}

func (m *manager) ListByTemplateName(ctx context.Context, templateName string) ([]TemplateRelease, error) {
	return m.dao.ListByTemplateName(ctx, templateName)
}

func (m *manager) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (*TemplateRelease, error) {
	return m.dao.GetByTemplateNameAndRelease(ctx, templateName, release)
}
