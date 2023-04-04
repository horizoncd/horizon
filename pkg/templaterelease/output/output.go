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

	tmanager "github.com/horizoncd/horizon/pkg/manager"
	"golang.org/x/net/context"

	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/templaterepo"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

// Getter provides some functions for output
type Getter interface {
	// GetTemplateOutPut get horizon template's output for specified template release.
	GetTemplateOutPut(ctx context.Context, templateName, releaseName string) (string, error)
}

type getter struct {
	templateRepo       templaterepo.TemplateRepo
	templateMgr        tmanager.TemplateManager
	templateReleaseMgr tmanager.TemplateReleaseManager
}

const (
	// output yaml file path
	_outputsPath = "output/outputs.yaml"
)

var (
	ErrGitlab = errors.New("GitlabError")
	ErrFormat = errors.New("FileFormatError")
)

func NewOutPutGetter(ctx context.Context, repo templaterepo.TemplateRepo, m *managerparam.Manager) (Getter, error) {
	return &getter{
		templateRepo:       repo,
		templateMgr:        m.TemplateMgr,
		templateReleaseMgr: m.TemplateReleaseManager,
	}, nil
}

func (g *getter) GetTemplateOutPut(ctx context.Context,
	templateName, releaseName string) (string, error) {
	const op = "template output getter: getTemplateOutPut"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return "", err
	}

	chart, err := g.templateRepo.GetChart(tr.ChartName, tr.ChartVersion, tr.LastSyncAt)
	if err != nil {
		return "", err
	}

	for _, file := range chart.Files {
		if file.Name == _outputsPath {
			return string(file.Data), nil
		}
	}

	return "", nil
}
