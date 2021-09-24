package template

import (
	"context"
	"net/http"
	"testing"

	gitlabctlmock "g.hz.netease.com/horizon/mock/controller/gitlab"
	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	trmock "g.hz.netease.com/horizon/pkg/templaterelease/mock"
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

func Test(t *testing.T) {
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
	gitlabLib.EXPECT().GetFile(ctx, templateGitlabProject, releaseName, schemaPath).Return(
		[]byte(jsonSchema), nil)

	ctl := &controller{
		templateReleaseMgr: templateReleaseMgr,
		gitlabCtl:          gitlabCtl,
	}

	// release exists
	b, err := ctl.GetTemplateSchema(ctx, templateName, releaseName)
	assert.Nil(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, jsonSchema, string(b))

	// release not exists
	b, err = ctl.GetTemplateSchema(ctx, templateName, "release-not-exists")
	assert.Nil(t, b)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
	assert.Equal(t, string(ErrCodeReleaseNotFound), errors.Code(err))
}
