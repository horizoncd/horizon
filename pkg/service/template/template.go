package template

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"g.hz.netease.com/horizon/pkg/dao/templaterelease"
	gitlabsvc "g.hz.netease.com/horizon/pkg/service/gitlab"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

var (
	Service = newTemplateService()
)

// Interface provides some functions for templates
type Interface interface {
	// GetTemplateSchema get schema for specified template release
	GetTemplateSchema(ctx context.Context, templateName, releaseName string) (*Schemas, error)
}

type Schemas struct {
	CD *Schema
	CI *Schema
}

type Schema struct {
	JSONSchema map[string]interface{}
	UISchema   map[string]interface{}
}

const (
	// json schema file path
	_ciSchemaPath = "schema/ci.schema.json"
	_cdSchemaPath = "schema/cd.schema.json"
	// ui schema file path
	_ciUISchemaPath = "schema/ci.ui.schema.json"
	_cdUISchemaPath = "schema/cd.ui.schema.json"

	// ErrCodeReleaseNotFound  ReleaseNotFound error code
	_errCodeReleaseNotFound = errors.ErrorCode("ReleaseNotFound")
)

type templateService struct {
	gitlabFactory      gitlabsvc.Factory
	templateReleaseMgr templaterelease.Manager
}

func newTemplateService() Interface {
	return &templateService{
		gitlabFactory:      gitlabsvc.Fty,
		templateReleaseMgr: templaterelease.Mgr,
	}
}

func (t *templateService) GetTemplateSchema(ctx context.Context,
	templateName, releaseName string) (_ *Schemas, err error) {
	const op = "template git: getTemplateSchema"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	tr, err := t.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return nil, errors.E(op, http.StatusNotFound, _errCodeReleaseNotFound,
			fmt.Sprintf("the release %v of template %v is not found", releaseName, templateName))
	}
	gitlabLib, err := t.gitlabFactory.GetByName(ctx, tr.GitlabName)
	if err != nil {
		return nil, err
	}

	// 1. read file concurrently
	var err1, err2, err3, err4 error
	var ciSchemaBytes, cdSchemaBytes, ciUISchemaBytes, cdUISchemaBytes []byte
	var wgReadFile sync.WaitGroup
	wgReadFile.Add(4)
	readFile := func(b *[]byte, err *error, filePath string) {
		defer wgReadFile.Done()
		bytes, e := gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, filePath)
		*b = bytes
		*err = e
	}
	go readFile(&ciSchemaBytes, &err1, _ciSchemaPath)
	go readFile(&cdSchemaBytes, &err2, _cdSchemaPath)
	go readFile(&ciUISchemaBytes, &err3, _ciUISchemaPath)
	go readFile(&cdUISchemaBytes, &err4, _cdUISchemaPath)
	wgReadFile.Wait()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return nil, err
		}
	}

	// 2. unmarshal concurrently
	var ciSchema, cdSchema, ciUISchema, cdUISchema map[string]interface{}
	var wgUnmarshal sync.WaitGroup
	wgUnmarshal.Add(4)
	unmarshal := func(b []byte, m *map[string]interface{}, err *error) {
		defer wgUnmarshal.Done()
		if e := json.Unmarshal(b, &m); e != nil {
			*err = e
		}
	}
	go unmarshal(ciSchemaBytes, &ciSchema, &err1)
	go unmarshal(cdSchemaBytes, &cdSchema, &err2)
	go unmarshal(ciUISchemaBytes, &ciUISchema, &err3)
	go unmarshal(cdUISchemaBytes, &cdUISchema, &err4)
	wgUnmarshal.Wait()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return nil, err
		}
	}

	return &Schemas{
		CD: &Schema{
			JSONSchema: cdSchema,
			UISchema:   cdUISchema,
		},
		CI: &Schema{
			JSONSchema: ciSchema,
			UISchema:   ciUISchema,
		},
	}, nil
}
