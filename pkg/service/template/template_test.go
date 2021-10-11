package template

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	trmock "g.hz.netease.com/horizon/mock/pkg/dao/templaterelease"
	gitlablibmock "g.hz.netease.com/horizon/mock/pkg/lib/gitlab"
	gitlabsvcmock "g.hz.netease.com/horizon/mock/pkg/service/gitlab"
	trmodels "g.hz.netease.com/horizon/pkg/dao/templaterelease"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx                   = context.Background()
	gitlabName            = "control"
	templateName          = "javaapp"
	releaseName           = "v1.0.0"
	templateGitlabProject = "helm-template/javaapp"
)

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	gitlabFty := gitlabsvcmock.NewMockFactory(mockCtl)
	gitlabLib := gitlablibmock.NewMockInterface(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)

	gitlabFty.EXPECT().GetByName(ctx, gitlabName).Return(gitlabLib, nil)

	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		releaseName).Return(&trmodels.TemplateRelease{
		Model: gorm.Model{
			ID: 1,
		},
		TemplateName:  templateName,
		Name:          releaseName,
		GitlabName:    gitlabName,
		GitlabProject: templateGitlabProject,
	}, nil)

	templateReleaseMgr.EXPECT().GetByTemplateNameAndRelease(ctx, templateName,
		"release-not-exists").Return(nil, nil)

	jsonSchema := `{"type": "object"}`
	var jsonSchemaMap map[string]interface{}
	_ = json.Unmarshal([]byte(jsonSchema), &jsonSchemaMap)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _ciSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _cdSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _ciUISchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _cdUISchemaPath).Return(
		[]byte(jsonSchema), nil)

	svc := &templateService{
		templateReleaseMgr: templateReleaseMgr,
		gitlabFactory:      gitlabFty,
	}

	// release exists
	schema, err := svc.GetTemplateSchema(ctx, templateName, releaseName)
	assert.Nil(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, jsonSchemaMap, schema.CD.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.CD.UISchema)
	assert.Equal(t, jsonSchemaMap, schema.CI.JSONSchema)
	assert.Equal(t, jsonSchemaMap, schema.CI.UISchema)

	// release not exists
	schema, err = svc.GetTemplateSchema(ctx, templateName, "release-not-exists")
	assert.Nil(t, schema)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
	assert.Equal(t, string(_errCodeReleaseNotFound), errors.Code(err))
}
