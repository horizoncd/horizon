// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	templatemock "github.com/horizoncd/horizon/mock/pkg/template/manager"
	templatereleasemock "github.com/horizoncd/horizon/mock/pkg/templaterelease/manager"
	repomock "github.com/horizoncd/horizon/mock/pkg/templaterepo"
	trm "github.com/horizoncd/horizon/pkg/templaterelease/models"
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

	tm := time.Now()
	tr := &trm.TemplateRelease{
		TemplateName: templateName,
		ChartVersion: releaseName,
		ChartName:    templateName,
		LastSyncAt:   tm,
	}

	// 2. test gitlab get file ok
	outputSchemaStr := "domain: s3.mockserver.org"
	templatereleaseMockMgr.EXPECT().GetByTemplateNameAndRelease(gomock.Any(),
		templateName, releaseName).Return(tr, nil)
	repoMock.EXPECT().GetChart(templateName, releaseName, tm).
		Return(&chart.Chart{Files: []*chart.File{{Name: _outputsPath, Data: []byte(outputSchemaStr)}}}, nil).Times(1)

	outputstr, err = outputGetter.GetTemplateOutPut(context.TODO(), templateName, releaseName)
	assert.Nil(t, err)
	assert.Equal(t, outputSchemaStr, outputstr)
}
