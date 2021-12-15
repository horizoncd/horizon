package output

import (
	"errors"
	"testing"

	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	templatereleasemock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	trm "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestGeTemplateOutPut(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gitlabmockLib := gitlablibmock.NewMockInterface(mockCtrl)
	templatereleasemock := templatereleasemock.NewMockManager(mockCtrl)

	var outputGetter Getter = &getter{
		gitlabLib:          gitlabmockLib,
		templateReleaseMgr: templatereleasemock,
	}

	templateName := "java"
	releaseName := "v1.0.0"
	retErr := errors.New("get template release error")
	templatereleasemock.EXPECT().GetByTemplateNameAndRelease(gomock.Any(),
		templateName, releaseName).Return(nil, retErr)

	// 1. test GetByTemplateNameAndRelease error
	outputstr, err := outputGetter.GetTemplateOutPut(context.TODO(), templateName, releaseName)
	assert.NotNil(t, err)
	assert.Equal(t, err, retErr)
	assert.Equal(t, outputstr, "")

	gitProject := "horizon.github.com"
	tr := &trm.TemplateRelease{
		Name:          templateName,
		GitlabProject: gitProject,
	}

	// 2. test gitlab get file ok
	outputSchemaStr := "domain: s3.mockserver.org"
	templatereleasemock.EXPECT().GetByTemplateNameAndRelease(gomock.Any(),
		templateName, releaseName).Return(tr, nil)
	gitlabmockLib.EXPECT().GetFile(gomock.Any(),
		gitProject, templateName, _outputsPath).Return([]byte(outputSchemaStr), nil).Times(1)

	outputstr, err = outputGetter.GetTemplateOutPut(context.TODO(), templateName, releaseName)
	assert.Nil(t, err)
	assert.Equal(t, outputstr, outputSchemaStr)
}
