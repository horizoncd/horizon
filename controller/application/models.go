package application

import (
	"g.hz.netease.com/horizon/pkg/application/models"
)

// Base holds the parameters which can be updated of an application
type Base struct {
	Description   string                 `json:"description"`
	Priority      string                 `json:"priority"`
	Template      *Template              `json:"template"`
	Git           *Git                   `json:"git"`
	TemplateInput map[string]interface{} `json:"templateInput"`
	PipelineInput map[string]interface{} `json:"pipelineInput"`
}

// CreateApplicationRequest holds the parameters required to create an application
type CreateApplicationRequest struct {
	Base

	Name    string `json:"name"`
	GroupID uint   `json:"groupID"`
}

// UpdateApplicationRequest holds the parameters required to update an application
type UpdateApplicationRequest struct {
	Base
}

type GetApplicationResponse struct {
	CreateApplicationRequest
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
func (m *CreateApplicationRequest) toApplicationModel() *models.Application {
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

// toApplicationModel transfer UpdateApplicationRequest to models.Application
func (m *UpdateApplicationRequest) toApplicationModel() *models.Application {
	return &models.Application{
		Description:     m.Description,
		Priority:        models.Priority(m.Priority),
		GitURL:          m.Git.URL,
		GitSubfolder:    m.Git.Subfolder,
		GitBranch:       m.Git.Branch,
		Template:        m.Template.Name,
		TemplateRelease: m.Template.Release,
	}
}

// ofApplicationModel transfer models.Application, templateInput, pipelineInput to GetApplicationResponse
func ofApplicationModel(app *models.Application,
	templateInput, pipelineInput map[string]interface{}) *GetApplicationResponse {
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
				TemplateInput: templateInput,
				PipelineInput: pipelineInput,
			},
			Name:    app.Name,
			GroupID: app.GroupID,
		},
	}
}
