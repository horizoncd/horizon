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

package template

import (
	"context"
	"regexp"
	"sort"
	"time"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	tmodels "github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	trschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
)

type CreateTemplateRequest struct {
	// when creating template, user must create a release,
	// otherwise, it's a useless template.
	CreateReleaseRequest `json:"release"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Repository           string `json:"repository"`
	Type                 string `json:"type"`
	OnlyOwner            bool   `json:"onlyOwner"`
}

func (c *CreateTemplateRequest) toTemplateModel(ctx context.Context) (*tmodels.Template, error) {
	if c.Repository == "" {
		return nil, perror.Wrap(herrors.ErrTemplateParamInvalid,
			"Repository is empty")
	}
	if !checkIfNameValid(c.Name) {
		return nil, perror.Wrap(herrors.ErrTemplateParamInvalid,
			"Name starts with a letter and consists of an "+
				"alphanumeric underscore with a maximum of 63 characters")
	}
	t := &tmodels.Template{
		Name:        c.Name,
		Description: c.Description,
		Repository:  c.Repository,
		OnlyOwner:   &c.OnlyOwner,
		Type:        c.Type,
	}
	return t, nil
}

type CreateReleaseRequest struct {
	Name        string `json:"name"`
	Recommended bool   `json:"recommended"`
	Description string `json:"description"`
	OnlyOwner   bool   `json:"onlyOwner"`
}

func (c *CreateReleaseRequest) toReleaseModel(ctx context.Context,
	template *tmodels.Template) (*trmodels.TemplateRelease, error) {
	t := &trmodels.TemplateRelease{
		Name:         c.Name,
		TemplateName: template.Name,
		ChartName:    template.ChartName,
		Description:  c.Description,
		Recommended:  &c.Recommended,
		OnlyOwner:    &c.OnlyOwner,
	}

	return t, nil
}

type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	OnlyOwner   bool   `json:"onlyOwner"`
	WithoutCI   bool   `json:"withoutCI"`
}

func (c *UpdateTemplateRequest) toTemplateModel(ctx context.Context) (*tmodels.Template, error) {
	if c.Repository == "" {
		return nil, perror.Wrap(herrors.ErrTemplateParamInvalid,
			"Repository is empty")
	}
	if !checkIfNameValid(c.Name) {
		return nil, perror.Wrap(herrors.ErrTemplateParamInvalid,
			"Name starts with a letter and consists of an "+
				"alphanumeric underscore with a maximum of 63 characters")
	}

	t := &tmodels.Template{
		Name:        c.Name,
		Description: c.Description,
		Repository:  c.Repository,
		OnlyOwner:   &c.OnlyOwner,
		WithoutCI:   c.WithoutCI,
	}
	return t, nil
}

type UpdateReleaseRequest struct {
	Name        string `json:"name"`
	Recommended *bool  `json:"recommended,omitempty"`
	Description string `json:"description"`
	OnlyOwner   bool   `json:"onlyOwner"`
}

func (c *UpdateReleaseRequest) toReleaseModel(ctx context.Context) (*trmodels.TemplateRelease, error) {
	tr := &trmodels.TemplateRelease{
		Name:        c.Name,
		Description: c.Description,
		Recommended: c.Recommended,
		OnlyOwner:   &c.OnlyOwner,
	}
	return tr, nil
}

type Template struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	ChartName   string    `json:"chartName"`
	Description string    `json:"description"`
	Repository  string    `json:"repository"`
	Releases    Releases  `json:"releases,omitempty"`
	FullPath    string    `json:"fullPath,omitempty"`
	GroupID     uint      `json:"group"`
	OnlyOwner   bool      `json:"onlyOwner"`
	WithoutCI   bool      `json:"withoutCI"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   uint      `json:"createdBy"`
	UpdatedBy   uint      `json:"updatedBy"`
}

func toTemplate(m *tmodels.Template) *Template {
	if m == nil {
		return nil
	}
	t := &Template{
		ID:          m.ID,
		Name:        m.Name,
		ChartName:   m.ChartName,
		Description: m.Description,
		Repository:  m.Repository,
		GroupID:     m.GroupID,
		WithoutCI:   m.WithoutCI,
		Type:        m.Type,
		CreatedAt:   m.Model.CreatedAt,
		UpdatedAt:   m.Model.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
	}
	if m.OnlyOwner != nil {
		t.OnlyOwner = *m.OnlyOwner
	} else {
		t.OnlyOwner = false
	}
	return t
}

type Templates []*Template

func toTemplates(mts []*tmodels.Template) Templates {
	templates := make(Templates, 0)
	for _, mt := range mts {
		t := toTemplate(mt)
		if t != nil {
			templates = append(templates, t)
		}
	}
	return templates
}

type Release struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	TemplateID     uint      `json:"templateID"`
	TemplateName   string    `json:"templateName"`
	ChartVersion   string    `json:"chartVersion"`
	Description    string    `json:"description"`
	Recommended    bool      `json:"recommended"`
	OnlyOwner      bool      `json:"onlyOwner"`
	CommitID       string    `json:"commitID"`
	SyncStatusCode uint8     `json:"syncStatusCode"`
	SyncStatus     string    `json:"syncStatus"`
	LastSyncAt     time.Time `json:"lastSyncAt"`
	FailedReason   string    `json:"failedReason"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	CreatedBy      uint      `json:"createdBy"`
	UpdatedBy      uint      `json:"updatedBy"`
}

type Releases []*Release

func (r Releases) Len() int {
	return len(r)
}

func (r Releases) Less(i, j int) bool {
	// recommended first
	if r[i].Recommended {
		return true
	}
	if r[j].Recommended {
		return false
	}
	return r[i].Name > r[j].Name
}

func (r Releases) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func toRelease(m *trmodels.TemplateRelease) *Release {
	if m == nil {
		return nil
	}
	tr := &Release{
		ID:             m.ID,
		Name:           m.Name,
		ChartVersion:   m.ChartVersion,
		Description:    m.Description,
		TemplateID:     m.Template,
		TemplateName:   m.TemplateName,
		SyncStatusCode: uint8(m.SyncStatus),
		LastSyncAt:     m.LastSyncAt,
		CommitID:       m.CommitID,
		FailedReason:   m.FailedReason,
		CreatedAt:      m.Model.CreatedAt,
		UpdatedAt:      m.Model.UpdatedAt,
		CreatedBy:      m.CreatedBy,
		UpdatedBy:      m.UpdatedBy,
	}
	switch trmodels.SyncStatus(tr.SyncStatusCode) {
	case trmodels.StatusSucceed:
		tr.SyncStatus = "Succeed"
	case trmodels.StatusUnknown:
		tr.SyncStatus = "Unknown"
	case trmodels.StatusFailed:
		tr.SyncStatus = "Failed"
	case trmodels.StatusOutOfSync:
		tr.SyncStatus = "OutOfSync"
	}
	if m.Recommended != nil {
		tr.Recommended = *m.Recommended
	}
	if m.OnlyOwner != nil {
		tr.OnlyOwner = *m.OnlyOwner
	} else {
		tr.OnlyOwner = false
	}
	return tr
}

func toReleases(trs []*trmodels.TemplateRelease) Releases {
	releases := make(Releases, 0)
	for _, tr := range trs {
		t := toRelease(tr)
		if t != nil {
			releases = append(releases, t)
		}
	}
	sort.Sort(releases)
	return releases
}

type Schemas struct {
	//
	Application *Schema `json:"application"`
	Pipeline    *Schema `json:"pipeline"`
}

type Schema struct {
	JSONSchema map[string]interface{} `json:"jsonSchema"`
	UISchema   map[string]interface{} `json:"uiSchema"`
}

func toSchemas(schemas *trschema.Schemas) *Schemas {
	if schemas == nil {
		return nil
	}
	return &Schemas{
		Application: &Schema{
			JSONSchema: schemas.Application.JSONSchema,
			UISchema:   schemas.Application.UISchema,
		},
		Pipeline: &Schema{
			JSONSchema: schemas.Pipeline.JSONSchema,
			UISchema:   schemas.Pipeline.UISchema,
		},
	}
}

func checkIfNameValid(name string) bool {
	if len(name) == 0 {
		return false
	}

	if len(name) > 40 {
		return false
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	pattern := regexp.MustCompile("^(([a-z][-a-z0-9_]*)?[a-z0-9])?$")
	return pattern.MatchString(name)
}
