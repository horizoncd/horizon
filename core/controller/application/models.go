package application

import (
	"g.hz.netease.com/horizon/pkg/dao/application"
)

// Base holds the parameters which can be updated of an application
type Base struct {
	Description   string         `json:"description"`
	Priority      string         `json:"priority"`
	Template      *Template      `json:"template"`
	Git           *Git           `json:"git"`
	TemplateInput *TemplateInput `json:"templateInput"`
}

type TemplateInput struct {
	CD map[string]interface{} `json:"cd"`
	CI map[string]interface{} `json:"ci"`
}

// CreateApplicationRequest holds the parameters required to create an application
type CreateApplicationRequest struct {
	Base

	Name string `json:"name"`
}

// UpdateApplicationRequest holds the parameters required to update an application
type UpdateApplicationRequest struct {
	Base
}

type GetApplicationResponse struct {
	CreateApplicationRequest

	GroupID uint `json:"groupID"`
}

// Template struct about template
type Template struct {
	Name    string `json:"name"`
	Release string `json:"release"`
}

// Git struct about git
type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch"`
}

// toApplicationModel transfer CreateApplicationRequest to models.Application
func (m *CreateApplicationRequest) toApplicationModel(groupID uint) *application.Application {
	return &application.Application{
		GroupID:         groupID,
		Name:            m.Name,
		Description:     m.Description,
		Priority:        application.Priority(m.Priority),
		GitURL:          m.Git.URL,
		GitSubfolder:    m.Git.Subfolder,
		GitBranch:       m.Git.Branch,
		Template:        m.Template.Name,
		TemplateRelease: m.Template.Release,
	}
}

// toApplicationModel transfer UpdateApplicationRequest to models.Application
func (m *UpdateApplicationRequest) toApplicationModel() *application.Application {
	return &application.Application{
		Description:     m.Description,
		Priority:        application.Priority(m.Priority),
		GitURL:          m.Git.URL,
		GitSubfolder:    m.Git.Subfolder,
		GitBranch:       m.Git.Branch,
		Template:        m.Template.Name,
		TemplateRelease: m.Template.Release,
	}
}

// ofApplicationModel transfer models.Application, templateInput, pipelineInput to GetApplicationResponse
func ofApplicationModel(app *application.Application,
	ci, cd map[string]interface{}) *GetApplicationResponse {
	return &GetApplicationResponse{
		CreateApplicationRequest: CreateApplicationRequest{
			Base: Base{
				Description: app.Description,
				Priority:    string(app.Priority),
				Template: &Template{
					Name:    app.Template,
					Release: app.TemplateRelease,
				},
				Git: &Git{
					URL:       app.GitURL,
					Subfolder: app.GitSubfolder,
					Branch:    app.GitBranch,
				},
				TemplateInput: &TemplateInput{
					CI: ci,
					CD: cd,
				},
			},
			Name: app.Name,
		},
		GroupID: app.GroupID,
	}
}
