package cluster

import (
	"github.com/horizoncd/horizon/core/common"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
)

type Cluster struct {
	*models.Cluster `json:",inline"`
	*TemplateInput  `json:"templateInput,omitempty"`
	Tags            tagmodels.TagsBasic `json:"tags,omitempty"`
}

type Clusterv2 struct {
	*models.Cluster `json:",inline"`
	*TemplateInput  `json:"templateInput,omitempty"`
	TemplateConfig  map[string]interface{} `json:"templateConfig,omitempty"`
	TemplateInfo    *code.TemplateInfo     `json:"-"`
	MergePatch      bool                   `json:"mergePatch"`
	BuildConfig     map[string]interface{} `json:"buildConfig"`
	Tags            tagmodels.TagsBasic    `json:"tags,omitempty"`
}

func (c *Clusterv2) toClusterModel(application *appmodels.Application) *models.Cluster {
	cluster := &models.Cluster{
		ApplicationID:   c.ApplicationID,
		Name:            c.Name,
		EnvironmentName: c.EnvironmentName,
		RegionName:      c.RegionName,
		Description:     c.Description,
		ExpireSeconds:   c.ExpireSeconds,
		Template:        c.Template,
		TemplateRelease: c.TemplateRelease,
		Status:          common.ClusterStatusCreating,
	}
	if cluster.Template == application.Template {
		cluster.GitURL = func() string {
			if c.GitURL == "" && application.GitURL != "" {
				return application.GitURL
			}
			// if URL is empty string, this means this cluster not depends on build from git
			return c.GitURL
		}()
		cluster.GitSubfolder = func() string {
			if c.GitSubfolder == "" {
				return application.GitSubfolder
			}
			return c.GitSubfolder
		}()
		cluster.GitRef = func() string {
			if c.GitRef == "" {
				return application.GitRef
			}
			return c.GitRef
		}()
		cluster.GitRefType = func() string {
			if c.GitRefType == "" {
				return application.GitRefType
			}
			return c.GitRefType
		}()
		cluster.Image = func() string {
			if c.Image == "" {
				return application.Image
			}
			return c.Image
		}()
	}
	return cluster
}
