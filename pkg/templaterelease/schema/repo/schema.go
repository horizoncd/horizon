package harbor

import (
	"context"
	"fmt"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
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
	repo templaterepo.TemplateRepo
}

func NewSchemaGetter(_ context.Context, repo templaterepo.TemplateRepo) *Getter {
	return &Getter{
		repo: repo,
	}
}

func (g *Getter) GetTemplateSchema(_ context.Context,
	templateName, releaseName string, params map[string]string) (*schema.Schemas, error) {
	chartPkg, err := g.repo.GetChart(templateName, releaseName)
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

	for name, file := range files {
		if file == nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("lack of template schema: %v", name))
		}
	}
	return schema.ParseFiles(params,
		files[_pipelineSchemaPath], files[_applicationSchemaPath],
		files[_pipelineUISchemaPath], files[_applicationUISchemaPath])
}
