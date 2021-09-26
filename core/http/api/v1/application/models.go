package application

import (
	"g.hz.netease.com/horizon/pkg/application/models"
)

type CreateApplication struct {
	GroupID       uint                   `json:"groupID"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Priority      string                 `json:"priority"`
	Template      *Template              `json:"template"`
	Git           *Git                   `json:"git"`
	TemplateInput map[string]interface{} `json:"templateInput"`
	PipelineInput map[string]interface{} `json:"pipelineInput"`
}

type Template struct {
	Name    string `json:"name"`
	Release string `json:"release"`
}

type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch"`
}

func (m *CreateApplication) toApplicationModel() *models.Application {
	return &models.Application{
		GroupID:         m.GroupID,
		Name:            m.Name,
		Description:     m.Description,
		Priority:        models.Priority(m.Priority),
		GitURL:          m.Git.URL,
		GitSubfolder:    m.Git.Subfolder,
		GitBranch:       m.Git.Branch,
		Template:        m.Template.Name,
		TemplateRelease: m.Template.Release,
	}
}
