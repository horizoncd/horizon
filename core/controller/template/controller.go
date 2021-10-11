package template

import (
	"context"

	"g.hz.netease.com/horizon/pkg/dao/template"
	"g.hz.netease.com/horizon/pkg/dao/templaterelease"
	templatesvc "g.hz.netease.com/horizon/pkg/service/template"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

var (
	Ctl = NewController()
)

type Controller interface {
	// ListTemplate list all template available
	ListTemplate(ctx context.Context) (Templates, error)
	// ListTemplateRelease list all releases of the specified template
	ListTemplateRelease(ctx context.Context, templateName string) (Releases, error)
	// GetTemplateSchema get schema for a template release
	GetTemplateSchema(ctx context.Context, templateName, releaseName string) (*Schemas, error)
}

type controller struct {
	templateMgr        template.Manager
	templateReleaseMgr templaterelease.Manager
	templateSvc        templatesvc.Interface
}

var _ Controller = (*controller)(nil)

// NewController initializes a new controller
func NewController() Controller {
	return &controller{
		templateMgr:        template.Mgr,
		templateReleaseMgr: templaterelease.Mgr,
		templateSvc:        templatesvc.Service,
	}
}

func (c *controller) ListTemplate(ctx context.Context) (_ Templates, err error) {
	const op = "template controller: listTemplate"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	templateModels, err := c.templateMgr.List(ctx)
	if err != nil {
		return nil, err
	}
	return toTemplates(templateModels), nil
}

func (c *controller) ListTemplateRelease(ctx context.Context, templateName string) (_ Releases, err error) {
	const op = "template controller: listTemplateRelease"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	templateReleaseModels, err := c.templateReleaseMgr.ListByTemplateName(ctx, templateName)
	if err != nil {
		return nil, err
	}
	return toReleases(templateReleaseModels), nil
}

func (c *controller) GetTemplateSchema(ctx context.Context, templateName, releaseName string) (_ *Schemas, err error) {
	const op = "template controller: getTemplateSchema"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	schemas, err := c.templateSvc.GetTemplateSchema(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}

	return toSchemas(schemas), nil
}
