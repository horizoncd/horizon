package harbor

import (
	"context"

	"g.hz.netease.com/horizon/pkg/param/managerparam"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/templaterepo"
)

const (
	// json schema file path
	_pipelineSchemaPath    = "schema/pipeline.schema.json"
	_applicationSchemaPath = "schema/application.schema.json"
	// ui schema file path
	_pipelineUISchemaPath    = "schema/pipeline.ui.schema.json"
	_applicationUISchemaPath = "schema/application.ui.schema.json"
)

type Getter struct {
	repo               templaterepo.TemplateRepo
	templateReleaseMgr trmanager.Manager
}

func NewSchemaGetter(_ context.Context, repo templaterepo.TemplateRepo,
	manager *managerparam.Manager) *Getter {
	return &Getter{
		repo:               repo,
		templateReleaseMgr: manager.TemplateReleaseManager,
	}
}

func (g *Getter) GetTemplateSchema(ctx context.Context,
	templateName, releaseName string, params map[string]string) (*schema.Schemas, error) {
	release, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}
	chartPkg, err := g.repo.GetChart(release.ChartName, release.ChartVersion, release.LastSyncAt)
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{
		_pipelineSchemaPath:      nil,
		_applicationSchemaPath:   nil,
		_pipelineUISchemaPath:    nil,
		_applicationUISchemaPath: nil,
	}

	for _, file := range chartPkg.Files {
		if t, ok := files[file.Name]; ok && t == nil {
			files[file.Name] = file.Data
		}
	}

	return schema.ParseFiles(params,
		files[_pipelineSchemaPath], files[_applicationSchemaPath],
		files[_pipelineUISchemaPath], files[_applicationUISchemaPath])
}
