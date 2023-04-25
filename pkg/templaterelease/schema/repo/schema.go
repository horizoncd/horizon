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

package harbor

import (
	"context"

	"github.com/horizoncd/horizon/pkg/param/managerparam"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	"github.com/horizoncd/horizon/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/templaterepo"
)

const (
	// json schema file path
	_pipelineSchemaPath    = "schema/pipeline.schema.json"
	_applicationSchemaPath = "schema/application.schema.json"
	// ui schema file path
	_pipelineUISchemaPath    = "schema/pipeline.ui.schema.json"
	_applicationUISchemaPath = "schema/application.ui.schema.json"
)

type Getter struct {
	repo               templaterepo.TemplateRepo
	templateReleaseMgr trmanager.Manager
}

func NewSchemaGetter(_ context.Context, repo templaterepo.TemplateRepo,
	manager *managerparam.Manager) *Getter {
	return &Getter{
		repo:               repo,
		templateReleaseMgr: manager.TemplateReleaseManager,
	}
}

func (g *Getter) GetTemplateSchema(ctx context.Context,
	templateName, releaseName string, params map[string]string) (*schema.Schemas, error) {
	release, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}
	chartPkg, err := g.repo.GetChart(release.ChartName, release.ChartVersion, release.LastSyncAt)
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{
		_pipelineSchemaPath:      nil,
		_applicationSchemaPath:   nil,
		_pipelineUISchemaPath:    nil,
		_applicationUISchemaPath: nil,
	}

	for _, file := range chartPkg.Files {
		if t, ok := files[file.Name]; ok && t == nil {
			files[file.Name] = file.Data
		}
	}

	return schema.ParseFiles(params,
		files[_pipelineSchemaPath], files[_applicationSchemaPath],
		files[_pipelineUISchemaPath], files[_applicationUISchemaPath])
}
