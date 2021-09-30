package template

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	gitlabctlmock "g.hz.netease.com/horizon/mock/controller/gitlab"
	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	tmock "g.hz.netease.com/horizon/mock/pkg/template"
	trmock "g.hz.netease.com/horizon/mock/pkg/templaterelease"
	"g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/util/errors"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"gorm.io/gorm"
)

var (
	ctx                   = context.Background()
	gitlabName            = "control"
	templateName          = "javaapp"
	releaseName           = "v1.0.0"
	templateGitlabProject = "helm-template/javaapp"
)

func TestList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	templateMgr := tmock.NewMockManager(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)

	templateMgr.EXPECT().List(ctx).Return([]models.Template{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "javaapp",
		}, {
			Model: gorm.Model{
				ID: 2,
			},
			Name: "tomcat",
		},
	}, nil)

	templateReleaseMgr.EXPECT().ListByTemplateName(ctx, "javaapp").
		Return([]trmodels.TemplateRelease{
			{
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.0",
				Recommended:  false,
			}, {
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.1",
				Recommended:  true,
			}, {
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.2",
				Recommended:  false,
			},
		}, nil)

	ctl := &controller{
		templateReleaseMgr: templateReleaseMgr,
		templateMgr:        templateMgr,
	}

	templates, err := ctl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, "javaapp", templates[0].Name)
	assert.Equal(t, "tomcat", templates[1].Name)

	templateReleases, err := ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(templateReleases))
	assert.Equal(t, "v1.0.1", templateReleases[0].Name)
	assert.Equal(t, "v1.0.2", templateReleases[1].Name)
	assert.Equal(t, "v1.0.0", templateReleases[2].Name)
}

func TestGetSchema(t *testing.T) {
	mockCtl := gomock.NewController(t)
	gitlabCtl := gitlabctlmock.NewMockController(mockCtl)
	gitlabLib := gitlablibmock.NewMockInterface(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)

	gitlabCtl.EXPECT().GetByName(ctx, gitlabName).Return(gitlabLib, nil)

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
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _cdSchemaPath).Return(
		[]byte(jsonSchema), nil)
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, _ciSchemaPath).Return(
		[]byte(jsonSchema), nil)

	ctl := &controller{
		templateReleaseMgr: templateReleaseMgr,
		gitlabCtl:          gitlabCtl,
	}

	// release exists
	schema, err := ctl.GetTemplateSchema(ctx, templateName, releaseName)
	assert.Nil(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, jsonSchemaMap, schema.CD)
	assert.Equal(t, jsonSchemaMap, schema.CI)

	// release not exists
	schema, err = ctl.GetTemplateSchema(ctx, templateName, "release-not-exists")
	assert.Nil(t, schema)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
	assert.Equal(t, string(_errCodeReleaseNotFound), errors.Code(err))
}
