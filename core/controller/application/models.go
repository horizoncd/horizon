package application

import (
	"time"

	"g.hz.netease.com/horizon/pkg/application/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
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
	Application map[string]interface{} `json:"application"`
	Pipeline    map[string]interface{} `json:"pipeline"`
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

	GroupID   uint      `json:"groupID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Template struct about template
type Template struct {
	Name               string `json:"name"`
	Release            string `json:"release"`
	RecommendedRelease string `json:"recommendedRelease,omitempty"`
}

// Git struct about git
type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch"`
}

// toApplicationModel transfer CreateApplicationRequest to models.Application
func (m *CreateApplicationRequest) toApplicationModel(groupID uint) *models.Application {
	return &models.Application{
		GroupID:         groupID,
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
func ofApplicationModel(app *models.Application, trs []*trmodels.TemplateRelease,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) *GetApplicationResponse {
	var recommendedRelease string
	for _, tr := range trs {
		if tr.Recommended {
			recommendedRelease = tr.Name
		}
	}
	return &GetApplicationResponse{
		CreateApplicationRequest: CreateApplicationRequest{
			Base: Base{
				Description: app.Description,
				Priority:    string(app.Priority),
				Template: &Template{
					Name:               app.Template,
					Release:            app.TemplateRelease,
					RecommendedRelease: recommendedRelease,
				},
				Git: &Git{
					URL:       app.GitURL,
					Subfolder: app.GitSubfolder,
					Branch:    app.GitBranch,
				},
				TemplateInput: &TemplateInput{
					Application: applicationJSONBlob,
					Pipeline:    pipelineJSONBlob,
				},
			},
			Name: app.Name,
		},
		GroupID:   app.GroupID,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}
}
