package schema

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	trmock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx                   = context.Background()
	templateName          = "javaapp"
	releaseName           = "v1.0.0"
	templateGitlabProject = "helm-template/javaapp"
)

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	gitlabLib := gitlablibmock.NewMockInterface(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)

	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		releaseName).Return(&trmodels.TemplateRelease{
		Model: gorm.Model{
			ID: 1,
		},
		TemplateName:  templateName,
		Name:          releaseName,
		GitlabProject: templateGitlabProject,
	}, nil)

	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		"release-not-exists").Return(nil, errors.E("", http.StatusNotFound))

	jsonSchema := `{"type": "object"}`
	var jsonSchemaMap map[string]interface{}
	_ = json.Unmarshal([]byte(jsonSchema), &jsonSchemaMap)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _pipelineSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _applicationSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _pipelineUISchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _applicationUISchemaPath).Return(
		[]byte(jsonSchema), nil)

	g := &getter{
		templateReleaseMgr: templateReleaseMgr,
		gitlabLib:          gitlabLib,
	}

	// release exists
	schema, err := g.GetTemplateSchema(ctx, templateName, releaseName)
	assert.Nil(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, jsonSchemaMap, schema.Application.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.Application.UISchema)
	assert.Equal(t, jsonSchemaMap, schema.Pipeline.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.Pipeline.UISchema)

	// release not exists
	schema, err = g.GetTemplateSchema(ctx, templateName, "release-not-exists")
	assert.Nil(t, schema)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
}
