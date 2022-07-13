package gitlab

import (
	"context"
	"sync"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	tmanager "g.hz.netease.com/horizon/pkg/template/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/schema"
	templateschemamanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const _gitlabName = "control"

const (
	// json schema file path
	_pipelineSchemaPath    = "schema/pipeline.schema.json"
	_applicationSchemaPath = "schema/application.schema.json"
	// ui schema file path
	_pipelineUISchemaPath    = "schema/pipeline.ui.schema.json"
	_applicationUISchemaPath = "schema/application.ui.schema.json"
)

// params
const (
	ClusterIDKey    string = "clusterID"
	ResourceTypeKey string = "resourceType"
)

type getter struct {
	gitlabLib            gitlablib.Interface
	templateMgr          tmanager.Manager
	templateReleaseMgr   trmanager.Manager
	templateSchemaTagMgr templateschemamanager.Manager
}

func NewSchemaGetter(ctx context.Context, gitlabFty gitlabfty.Factory,
	manager *managerparam.Manager) (schema.Getter, error) {
	gitlabLib, err := gitlabFty.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &getter{
		gitlabLib:            gitlabLib,
		templateMgr:          manager.TemplateMgr,
		templateReleaseMgr:   manager.TemplateReleaseManager,
		templateSchemaTagMgr: manager.TemplateSchemaTagManager,
	}, nil
}

func (g *getter) GetTemplateSchema(ctx context.Context,
	templateName, releaseName string, params map[string]string) (_ *schema.Schemas, err error) {
	const op = "template schema getter: getTemplateSchema"
	defer wlog.Start(ctx, op).StopPrint()

	t, err := g.templateMgr.GetByName(ctx, templateName)
	tr, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}

	// 1. read file concurrently
	var err1, err2, err3, err4 error
	var pipelineSchemaBytes, applicationSchemaBytes, pipelineUISchemaBytes, applicationUISchemaBytes []byte
	var wgReadFile sync.WaitGroup
	wgReadFile.Add(4)
	readFile := func(b *[]byte, err *error, filePath string) {
		defer wgReadFile.Done()
		bts, e := g.gitlabLib.GetFile(ctx, t.Repository, tr.Name, filePath)
		*b = bts
		*err = e
	}
	go readFile(&pipelineSchemaBytes, &err1, _pipelineSchemaPath)
	go readFile(&applicationSchemaBytes, &err2, _applicationSchemaPath)
	go readFile(&pipelineUISchemaBytes, &err3, _pipelineUISchemaPath)
	go readFile(&applicationUISchemaBytes, &err4, _applicationUISchemaPath)
	wgReadFile.Wait()
	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return nil, err
		}
	}
	return schema.ParseFiles(params,
		pipelineSchemaBytes, applicationSchemaBytes,
		pipelineUISchemaBytes, applicationUISchemaBytes)
}
