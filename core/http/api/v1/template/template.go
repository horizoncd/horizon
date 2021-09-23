package template

import (
	"sort"

	"g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
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
