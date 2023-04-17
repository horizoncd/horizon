package application

import (
	"time"

	"github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
)

// Base holds the parameters which can be updated of an application.
type Base struct {
	Description   string          `json:"description"`
	Priority      string          `json:"priority"`
	Template      *Template       `json:"template"`
	Git           *codemodels.Git `json:"git"`
	TemplateInput *TemplateInput  `json:"templateInput"`
}

type TemplateInput struct {
	Application map[string]interface{} `json:"application"`
	Pipeline    map[string]interface{} `json:"pipeline"`
}

// CreateApplicationRequest holds the parameters required to create an application.
type CreateApplicationRequest struct {
	Base

	Name         string            `json:"name"`
	ExtraMembers map[string]string `json:"extraMembers"`
}

// UpdateApplicationRequest holds the parameters required to update an application.
type UpdateApplicationRequest struct {
	Base
}

type GetApplicationResponse struct {
	CreateApplicationRequest
	FullPath  string    `json:"fullPath"`
	ID        uint      `json:"id"`
	GroupID   uint      `json:"groupID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListApplicationResponse struct {
	FullPath  string    `json:"fullPath"`
	FullName  string    `json:"fullName"`
	Name      string    `json:"name"`
	ID        uint      `json:"id"`
	GroupID   uint      `json:"groupID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Template struct about template.
type Template struct {
	Name               string `json:"name"`
	Release            string `json:"release"`
	RecommendedRelease string `json:"recommendedRelease,omitempty"`
}

// toApplicationModel transfer CreateApplicationRequest to models.Application.
func (m *CreateApplicationRequest) toApplicationModel(groupID uint) *models.Application {
	return &models.Application{
		GroupID:         groupID,
		Name:            m.Name,
		Description:     m.Description,
		Priority:        models.Priority(m.Priority),
		GitURL:          m.Git.URL,
		GitSubfolder:    m.Git.Subfolder,
		GitRefType:      m.Git.RefType(),
		GitRef:          m.Git.Ref(),
		Template:        m.Template.Name,
		TemplateRelease: m.Template.Release,
	}
}

// toApplicationModel transfer UpdateApplicationRequest to models.Application.
func (m *UpdateApplicationRequest) toApplicationModel(appExistsInDB *models.Application) *models.Application {
	application := &models.Application{
		Description:     appExistsInDB.Description,
		Priority:        appExistsInDB.Priority,
		GitURL:          appExistsInDB.GitURL,
		GitSubfolder:    appExistsInDB.GitSubfolder,
		GitRef:          appExistsInDB.GitRef,
		GitRefType:      appExistsInDB.GitRefType,
		Template:        appExistsInDB.Template,
		TemplateRelease: appExistsInDB.TemplateRelease,
	}
	application.Description = m.Description
	if m.Priority != "" {
		application.Priority = models.Priority(m.Priority)
	}
	if m.Git != nil {
		if m.Git.URL != "" {
			application.GitURL = m.Git.URL
		}
		application.GitRefType = m.Git.RefType()
		application.GitRef = m.Git.Ref()
		application.GitSubfolder = m.Git.Subfolder
	}
	if m.Template != nil {
		if m.Template.Name != "" {
			application.Template = m.Template.Name
		}
		if m.Template.Release != "" {
			application.TemplateRelease = m.Template.Release
		}
	}

	return application
}

// ofApplicationModel transfer models.Application, templateInput, pipelineInput to GetApplicationResponse.
func ofApplicationModel(app *models.Application, fullPath string, trs []*trmodels.TemplateRelease,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{},
) *GetApplicationResponse {
	var recommendedRelease string
	for _, tr := range trs {
		if *tr.Recommended {
			recommendedRelease = tr.Name
		}
	}

	resp := &GetApplicationResponse{
		CreateApplicationRequest: CreateApplicationRequest{
			Base: Base{
				Description: app.Description,
				Priority:    string(app.Priority),
				Template: &Template{
					Name:               app.Template,
					Release:            app.TemplateRelease,
					RecommendedRelease: recommendedRelease,
				},
				Git: codemodels.NewGit(app.GitURL, app.GitSubfolder, app.GitRefType, app.GitRef),
				TemplateInput: &TemplateInput{
					Application: applicationJSONBlob,
					Pipeline:    pipelineJSONBlob,
				},
			},
			Name: app.Name,
		},
		FullPath:  fullPath,
		ID:        app.ID,
		GroupID:   app.GroupID,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}
	return resp
}
