package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"text/template"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschemamanager "g.hz.netease.com/horizon/pkg/templateschema/manager"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/Masterminds/sprig"
)

const _gitlabName = "control"

// Getter provides some functions for template schema
type Getter interface {
	// GetTemplateSchema get schema for specified template release. todo(gjq) add cache
	GetTemplateSchema(ctx context.Context, templateName, releaseName string, params map[string]string) (*Schemas, error)
}

type Schemas struct {
	Application *Schema
	Pipeline    *Schema
}

type Schema struct {
	JSONSchema map[string]interface{}
	UISchema   map[string]interface{}
}

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
	ClusterIDKey string = "ClusterIDKey"
)

type getter struct {
	gitlabLib          gitlablib.Interface
	templateReleaseMgr manager.Manager
	templateSchemaMGr  templateschemamanager.Manager
}

func NewSchemaGetter(ctx context.Context, gitlabFty gitlabfty.Factory) (Getter, error) {
	gitlabLib, err := gitlabFty.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &getter{
		gitlabLib:          gitlabLib,
		templateReleaseMgr: manager.Mgr,
		templateSchemaMGr:  templateschemamanager.Mgr,
	}, nil
}

func (g *getter) GeneratorRenderParams(ctx context.Context, params map[string]string) (map[string]string, error) {
	clusterIDStr, ok := params[ClusterIDKey]
	if ok {
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
		if err != nil {
			return nil, err
		}
		tags, err := g.templateSchemaMGr.ListByClusterID(ctx, uint(clusterID))
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			params[tag.Key] = tag.Value
		}
		delete(params, ClusterIDKey)
	}
	return params, nil
}

func RenderFiles(params map[string]string, files ...[]byte) (retFiles [][]byte, _ error) {
	for _, file := range files {
		var b bytes.Buffer
		dotemplate := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(file)))
		err := dotemplate.ExecuteTemplate(&b, "", params)
		if err != nil {
			return nil, err
		}
		retFiles = append(retFiles, b.Bytes())
	}
	return retFiles, nil
}

func (g *getter) GetTemplateSchema(ctx context.Context,
	templateName, releaseName string, params map[string]string) (_ *Schemas, err error) {
	const op = "template schema getter: getTemplateSchema"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

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
		bytes, e := g.gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, filePath)
		*b = bytes
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

	// 2. get template schema tags and do template
	if len(params) != 0 {
		params, err = g.GeneratorRenderParams(ctx, params)
		if err != nil {
			return nil, err
		}
		readerSchemas, err := RenderFiles(params, pipelineSchemaBytes, applicationSchemaBytes)
		if err != nil {
			return nil, err
		}
		pipelineSchemaBytes = readerSchemas[0]
		applicationSchemaBytes = readerSchemas[1]
	}

	// 2. unmarshal concurrently
	var pipelineSchema, applicationSchema, pipelineUISchema, applicationUISchema map[string]interface{}
	var wgUnmarshal sync.WaitGroup
	wgUnmarshal.Add(4)
	unmarshal := func(b []byte, m *map[string]interface{}, err *error) {
		defer wgUnmarshal.Done()
		if e := json.Unmarshal(b, &m); e != nil {
			*err = e
		}
	}
	go unmarshal(pipelineSchemaBytes, &pipelineSchema, &err1)
	go unmarshal(applicationSchemaBytes, &applicationSchema, &err2)
	go unmarshal(pipelineUISchemaBytes, &pipelineUISchema, &err3)
	go unmarshal(applicationUISchemaBytes, &applicationUISchema, &err4)
	wgUnmarshal.Wait()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return nil, err
		}
	}

	return &Schemas{
		Application: &Schema{
			JSONSchema: applicationSchema,
			UISchema:   applicationUISchema,
		},
		Pipeline: &Schema{
			JSONSchema: pipelineSchema,
			UISchema:   pipelineUISchema,
		},
	}, nil
}
