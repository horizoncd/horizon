package template

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"g.hz.netease.com/horizon/controller/gitlab"
	"g.hz.netease.com/horizon/pkg/template"
	"g.hz.netease.com/horizon/pkg/templaterelease"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

const (
	// schemaPath template input.schema.json file path
	_cdSchemaPath = "schema/cd.schema.json"
	_ciSchemaPath = "schema/ci.schema.json"

	// ErrCodeReleaseNotFound  ReleaseNotFound error code
	_errCodeReleaseNotFound = errors.ErrorCode("ReleaseNotFound")
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
	GetTemplateSchema(ctx context.Context, templateName, releaseName string) (*Schema, error)
}

type controller struct {
	templateMgr        template.Manager
	templateReleaseMgr templaterelease.Manager
	gitlabCtl          gitlab.Controller
}

var _ Controller = (*controller)(nil)

// NewController initializes a new controller
func NewController() Controller {
	return &controller{
		templateMgr:        template.Mgr,
		templateReleaseMgr: templaterelease.Mgr,
		gitlabCtl:          gitlab.Ctl,
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

func (c *controller) GetTemplateSchema(ctx context.Context, templateName, releaseName string) (_ *Schema, err error) {
	const op = "template controller: getTemplateSchema"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return nil, errors.E(op, http.StatusNotFound, _errCodeReleaseNotFound,
			fmt.Sprintf("the release %v of template %v is not found", releaseName, templateName))
	}
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, tr.GitlabName)
	if err != nil {
		return nil, err
	}

	var err1, err2 error
	var ciSchemaBytes, cdSchemaBytes []byte
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		cdSchemaBytes, err1 = gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, _cdSchemaPath)
	}()
	go func() {
		defer wg.Done()
		ciSchemaBytes, err2 = gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, _ciSchemaPath)
	}()
	wg.Wait()

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	var ciSchema, cdSchema map[string]interface{}
	if err := json.Unmarshal(ciSchemaBytes, &ciSchema); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cdSchemaBytes, &cdSchema); err != nil {
		return nil, err
	}

	return &Schema{
		CD: cdSchema,
		CI: ciSchema,
	}, nil
}
