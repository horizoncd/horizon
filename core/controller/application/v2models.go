package application

import (
	"time"

	"g.hz.netease.com/horizon/pkg/application/models"
	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
)

// CreateApplicationRequestV2 holds the parameters required to create an application
type CreateApplicationRequestV2 struct {
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	Priority       *string                 `json:"priority"`
	Git            *codemodels.Git         `json:"git"`
	BuildConfig    *map[string]interface{} `json:"buildConfig"`
	TemplateConfig *TemplateConfig         `json:"templateConfig"`

	// TODO(remove it): only for internal usage
	ExtraMembers map[string]string `json:"extraMembers"`
}

func (m *CreateApplicationRequestV2) toApplicationModel(groupID uint) *models.Application {
	return &models.Application{
		GroupID:     groupID,
		Name:        m.Name,
		Description: m.Description,
		Priority: func() models.Priority {
			if m.Priority != nil {
				return models.Priority(*m.Priority)
			}
			return ""
		}(),
		GitURL: func() string {
			if m.Git != nil {
				return m.Git.URL
			}
			return ""
		}(),
		GitSubfolder: func() string {
			if m.Git != nil {
				return m.Git.Subfolder
			}
			return ""
		}(),
		GitRefType: func() string {
			if m.Git != nil {
				return m.Git.RefType()
			}
			return ""
		}(),
		GitRef: func() string {
			if m.Git != nil {
				return m.Git.Ref()
			}
			return ""
		}(),
		Template:        m.TemplateConfig.Name,
		TemplateRelease: m.TemplateConfig.Release,
	}
}

type TemplateConfig struct {
	TemplateType  string                  `json:"type"`
	Name          string                  `json:"name"`
	Release       string                  `json:"release"`
	TemplateInput *map[string]interface{} `json:"templateInput"`
}

type CreateApplicationResponseV2 struct {
	ID        uint      `json:"id"`
	FullPath  string    `json:"fullPath"`
	GroupID   uint      `json:"groupID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
