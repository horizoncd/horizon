package template

import (
	"context"
	"regexp"
	"sort"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	tmodels "g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	trschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
)

type CreateTemplateRequest struct {
	// when creating template, user must create a release,
	// otherwise, it's a useless template.
	CreateReleaseRequest `json:"release"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Repository           string `json:"repository"`
	OnlyAdmin            *bool  `json:"onlyAdmin"`
}

func (c *CreateTemplateRequest) toTemplateModel(ctx context.Context) (*tmodels.Template, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

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
	}
	onlyAdminFalse := false
	if user.IsAdmin() {
		t.OnlyAdmin = c.OnlyAdmin
	} else {
		t.OnlyAdmin = &onlyAdminFalse
	}
	return t, nil
}

type CreateReleaseRequest struct {
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	Recommended bool   `json:"recommended"`
	Description string `json:"description"`
	OnlyAdmin   *bool  `json:"onlyAdmin"`
}

func (c *CreateReleaseRequest) toReleaseModel(ctx context.Context,
	template *tmodels.Template) (*trmodels.TemplateRelease, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	t := &trmodels.TemplateRelease{
		Name:         c.Name,
		Tag:          c.Tag,
		TemplateName: template.Name,
		ChartName:    template.ChartName,
		Description:  c.Description,
		Recommended:  &c.Recommended,
	}

	onlyAdminFalse := false
	if user.IsAdmin() {
		t.OnlyAdmin = c.OnlyAdmin
	} else {
		t.OnlyAdmin = &onlyAdminFalse
	}
	return t, nil
}

type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	OnlyAdmin   *bool  `json:"onlyAdmin"`
}

func (c *UpdateTemplateRequest) toTemplateModel(ctx context.Context) (*tmodels.Template, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

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
	}
	if user.IsAdmin() {
		t.OnlyAdmin = c.OnlyAdmin
	}
	return t, nil
}

type UpdateReleaseRequest struct {
	Name        string `json:"name"`
	Recommended *bool  `json:"recommended,omitempty"`
	Description string `json:"description"`
	OnlyAdmin   *bool  `json:"onlyAdmin"`
}

func (c *UpdateReleaseRequest) toReleaseModel(ctx context.Context) (*trmodels.TemplateRelease, error) {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tr := &trmodels.TemplateRelease{
		Name:        c.Name,
		Description: c.Description,
		Recommended: c.Recommended,
	}
	if user.IsAdmin() {
		tr.OnlyAdmin = c.OnlyAdmin
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
	FullPath    string    `json:"fullpath,omitempty"`
	GroupID     uint      `json:"group"`
	OnlyAdmin   bool      `json:"onlyAdmin"`
	CreateAt    time.Time `json:"createAt"`
	UpdateAt    time.Time `json:"updateAt"`
	CreatedBy   uint      `json:"createBy"`
	UpdatedBy   uint      `json:"updateBy"`
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
		CreateAt:    m.Model.CreatedAt,
		UpdateAt:    m.Model.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
	}
	if m.OnlyAdmin != nil {
		t.OnlyAdmin = *m.OnlyAdmin
	}
	return t
}

type Templates []Template

func toTemplates(mts []*tmodels.Template) Templates {
	templates := make(Templates, 0)
	for _, mt := range mts {
		t := toTemplate(mt)
		if t != nil {
			templates = append(templates, *t)
		}
	}
	return templates
}

type Release struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Tag            string    `json:"tag"`
	TemplateID     uint      `json:"templateID"`
	ChartVersion   string    `json:"chartVersion"`
	Description    string    `json:"description"`
	Recommended    bool      `json:"recommended"`
	OnlyAdmin      bool      `json:"onlyAdmin"`
	CommitID       string    `json:"commitID"`
	SyncStatusCode uint8     `json:"syncStatusCode"`
	SyncStatus     string    `json:"syncStatus"`
	LastSyncAt     time.Time `json:"lastSyncAt"`
	FailedReason   string    `json:"failedReason"`
	CreateAt       time.Time `json:"createAt"`
	UpdateAt       time.Time `json:"updateAt"`
	CreatedBy      uint      `json:"createBy"`
	UpdatedBy      uint      `json:"updateBy"`
}

type Releases []Release

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
		Tag:            m.Tag,
		ChartVersion:   m.ChartVersion,
		Description:    m.Description,
		TemplateID:     m.Template,
		SyncStatusCode: uint8(m.SyncStatus),
		LastSyncAt:     m.LastSyncAt,
		CommitID:       m.CommitID,
		FailedReason:   m.FailedReason,
		CreateAt:       m.Model.CreatedAt,
		UpdateAt:       m.Model.UpdatedAt,
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
	if m.OnlyAdmin != nil {
		tr.OnlyAdmin = *m.OnlyAdmin
	}
	return tr
}

func toReleases(trs []*trmodels.TemplateRelease) Releases {
	releases := make(Releases, 0)
	for _, tr := range trs {
		t := toRelease(tr)
		if t != nil {
			releases = append(releases, *t)
		}
	}
	sort.Sort(releases)
	return releases
}

type Schemas struct {
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
