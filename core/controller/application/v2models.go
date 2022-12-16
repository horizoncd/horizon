package application

import (
	"time"

	"github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
)

type GetApplicationResponseV2 struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Priority    string          `json:"priority"`
	Git         *codemodels.Git `json:"git"`

	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
	Manifest       map[string]interface{}   `json:"manifest"`

	FullPath string `json:"fullPath"`
	GroupID  uint   `json:"groupID"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateOrUpdateApplicationRequestV2 holds the parameters required to create an application
type CreateOrUpdateApplicationRequestV2 struct {
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Priority       *string                  `json:"priority"`
	Git            *codemodels.Git          `json:"git"`
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`

	// TODO(remove it): only for internal usage
	ExtraMembers map[string]string `json:"extraMembers"`
}

func (req *CreateOrUpdateApplicationRequestV2) CreateToApplicationModel(groupID uint) *models.Application {
	return &models.Application{
		GroupID:     groupID,
		Name:        req.Name,
		Description: req.Description,
		Priority: func() models.Priority {
			if req.Priority != nil {
				return models.Priority(*req.Priority)
			}
			return ""
		}(),
		GitURL: func() string {
			if req.Git != nil {
				return req.Git.URL
			}
			return ""
		}(),
		GitSubfolder: func() string {
			if req.Git != nil {
				return req.Git.Subfolder
			}
			return ""
		}(),
		GitRefType: func() string {
			if req.Git != nil {
				return req.Git.RefType()
			}
			return ""
		}(),
		GitRef: func() string {
			if req.Git != nil {
				return req.Git.Ref()
			}
			return ""
		}(),
		Template: func() string {
			if req.TemplateInfo != nil {
				return req.TemplateInfo.Name
			}
			return ""
		}(),
		TemplateRelease: func() string {
			if req.TemplateInfo != nil {
				return req.TemplateInfo.Release
			}
			return ""
		}(),
	}
}

func (req *CreateOrUpdateApplicationRequestV2) UpdateToApplicationModel(
	appExistsInDB *models.Application) *models.Application {
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
	application.Description = req.Description
	if req.Priority != nil {
		application.Priority = models.Priority(*req.Priority)
	}
	if req.Git != nil {
		application.GitURL = req.Git.URL
		application.GitRefType = req.Git.RefType()
		application.GitRef = req.Git.Ref()
		application.GitSubfolder = req.Git.Subfolder
	}

	if req.TemplateInfo != nil {
		application.Template = req.TemplateInfo.Name
		application.TemplateRelease = req.TemplateInfo.Release
	}
	return application
}

type CreateApplicationResponseV2 struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Priority  string    `json:"priority"`
	FullPath  string    `json:"fullPath"`
	GroupID   uint      `json:"groupID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
