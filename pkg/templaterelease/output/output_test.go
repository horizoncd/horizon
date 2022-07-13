package output

import (
	"errors"
	"testing"

	templatemock "g.hz.netease.com/horizon/mock/pkg/template/manager"
	templatereleasemock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	repomock "g.hz.netease.com/horizon/mock/pkg/templaterepo"
	trm "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"helm.sh/helm/v3/pkg/chart"
)

func TestGeTemplateOutPut(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := repomock.NewMockTemplateRepo(mockCtrl)
	templateMockMgr := templatemock.NewMockManager(mockCtrl)
	templatereleaseMockMgr := templatereleasemock.NewMockManager(mockCtrl)

	var outputGetter Getter = &getter{
		templateRepo:       repoMock,
		templateMgr:        templateMockMgr,
		templateReleaseMgr: templatereleaseMockMgr,
	}

	templateName := "java"
	releaseName := "v1.0.0"
	retErr := errors.New("get template release error")
	templatereleaseMockMgr.EXPECT().GetByTemplateNameAndRelease(gomock.Any(),
		templateName, releaseName).Return(nil, retErr)

	// 1. test GetByTemplateNameAndRelease error
	outputstr, err := outputGetter.GetTemplateOutPut(context.TODO(), templateName, releaseName)
	assert.NotNil(t, err)
	assert.Equal(t, err, retErr)
	assert.Equal(t, outputstr, "")

	tr := &trm.TemplateRelease{
		TemplateName: templateName,
		Name:         releaseName,
		ChartName:    templateName,
	}

	// 2. test gitlab get file ok
	outputSchemaStr := "domain: s3.mockserver.org"
	templatereleaseMockMgr.EXPECT().GetByTemplateNameAndRelease(gomock.Any(),
		templateName, releaseName).Return(tr, nil)
	repoMock.EXPECT().GetChart(templateName, releaseName).
		Return(&chart.Chart{Files: []*chart.File{{Name: _outputsPath, Data: []byte(outputSchemaStr)}}}, nil).Times(1)

	outputstr, err = outputGetter.GetTemplateOutPut(context.TODO(), templateName, releaseName)
	assert.Nil(t, err)
	assert.Equal(t, outputSchemaStr, outputstr)
}
