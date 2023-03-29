package gitlab

import (
	"context"
	"sync"

	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	perror "github.com/horizoncd/horizon/pkg/errors"
	tmanager "github.com/horizoncd/horizon/pkg/template/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	"github.com/horizoncd/horizon/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

const (
	// json schema file path
	_pipelineSchemaPath    = "schema/pipeline.schema.json"
	_applicationSchemaPath = "schema/application.schema.json"
	// ui schema file path
	_pipelineUISchemaPath    = "schema/pipeline.ui.schema.json"
	_applicationUISchemaPath = "schema/application.ui.schema.json"
)

const (
	ClusterIDKey string = "clusterID"
)

type getter struct {
	gitlabLib          gitlablib.Interface
	templateMgr        tmanager.Manager
	templateReleaseMgr trmanager.Manager
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
		bts, e := g.gitlabLib.GetFile(ctx, t.Repository, tr.ChartVersion, filePath)
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
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return nil, err
			}
		}
	}
	return schema.ParseFiles(params,
		pipelineSchemaBytes, applicationSchemaBytes,
		pipelineUISchemaBytes, applicationUISchemaBytes)
}
