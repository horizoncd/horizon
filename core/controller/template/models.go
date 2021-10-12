package template

import (
	"sort"

	"g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	templatesvc "g.hz.netease.com/horizon/pkg/templaterelease/schema"
)

type Template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Templates []Template

func toTemplates(mts []models.Template) Templates {
	templates := make(Templates, 0)
	for _, mt := range mts {
		templates = append(templates, Template{
			Name:        mt.Name,
			Description: mt.Description,
		})
	}
	return templates
}

type Release struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Recommended bool   `json:"recommended"`
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

func toReleases(trs []trmodels.TemplateRelease) Releases {
	releases := make(Releases, 0)
	for _, tr := range trs {
		releases = append(releases, Release{
			Name:        tr.Name,
			Description: tr.Description,
			Recommended: tr.Recommended,
		})
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

func toSchemas(schemas *templatesvc.Schemas) *Schemas {
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
