package cluster

import (
	"time"

	"github.com/horizoncd/horizon/core/common"
	controllertag "github.com/horizoncd/horizon/core/controller/tag"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
)

type CreateClusterRequestV2 struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Priority    string               `json:"priority"`
	ExpireTime  string               `json:"expireTime"`
	Git         *codemodels.Git      `json:"git"`
	Tags        []*controllertag.Tag `json:"tags"`

	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`

	// TODO(tom): just for internal usage
	ExtraMembers map[string]string `json:"extraMembers"`
}

func (r *CreateClusterRequestV2) toClusterModel(application *appmodels.Application,
	er *envregionmodels.EnvironmentRegion, info *BuildTemplateInfo,
	expireSeconds uint,
) (*models.Cluster, []*tagmodels.Tag) {
	cluster := &models.Cluster{
		ApplicationID:   application.ID,
		Name:            r.Name,
		EnvironmentName: er.EnvironmentName,
		RegionName:      er.RegionName,
		Description:     r.Description,
		ExpireSeconds:   expireSeconds,
		// cluster provide git info or just use the application's git info
		GitURL: func() string {
			if r.Git == nil {
				return application.GitURL
			}
			if r.Git != nil && r.Git.URL == "" && application.GitURL != "" {
				return application.GitURL
			}
			// if URL is empty string, this means this cluster not depends on build from git
			return r.Git.URL
		}(),
		GitSubfolder: func() string {
			if r.Git == nil || r.Git.Subfolder == "" {
				return application.GitSubfolder
			}
			return r.Git.Subfolder
		}(),
		GitRef: func() string {
			if r.Git == nil {
				return application.GitRef
			}
			return r.Git.Ref()
		}(),
		GitRefType: func() string {
			if r.Git == nil {
				return application.GitRefType
			}
			return r.Git.RefType()
		}(),
		Template:        info.TemplateInfo.Name,
		TemplateRelease: info.TemplateInfo.Release,
		Status:          common.ClusterStatusCreating,
	}
	tags := make([]*tagmodels.Tag, 0)
	for _, tag := range r.Tags {
		tags = append(tags, &tagmodels.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return cluster, tags
}

type CreateClusterResponseV2 struct {
	ID            uint         `json:"id"`
	Name          string       `json:"name"`
	FullPath      string       `json:"fullPath"`
	ApplicationID uint         `json:"applicationID"`
	Scope         *Scope       `json:"scope"`
	Application   *Application `json:"application"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type UpdateClusterRequestV2 struct {
	// basic infos
	Description string `json:"description"`
	Priority    string `json:"priority"`
	ExpireTime  string `json:"expireTime"`

	// env and region info (can only be modified after cluster freed)
	Environment *string `json:"environment"`
	Region      *string `json:"region"`

	// source info
	Git *codemodels.Git `json:"git"`

	// git config info
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
}

func (r *UpdateClusterRequestV2) toClusterModel(cluster *models.Cluster, expireSeconds uint, environmentName,
	regionName, templateName, templateRelease string,
) *models.Cluster {
	var gitURL, gitSubFolder, gitRef, gitRefType string
	if r.Git != nil {
		gitURL, gitSubFolder, gitRefType, gitRef = r.Git.URL,
			r.Git.Subfolder, r.Git.RefType(), r.Git.Ref()
	} else {
		gitURL = cluster.GitURL
		gitSubFolder = cluster.GitSubfolder
		gitRefType = cluster.GitRefType
		gitRef = cluster.GitRef
	}
	return &models.Cluster{
		EnvironmentName: environmentName,
		RegionName:      regionName,
		Description:     r.Description,
		ExpireSeconds:   expireSeconds,
		GitURL:          gitURL,
		GitSubfolder:    gitSubFolder,
		GitRef:          gitRef,
		GitRefType:      gitRefType,
		Template:        templateName,
		TemplateRelease: templateRelease,
	}
}

type GetClusterResponseV2 struct {
	// basic infos
	ID              uint                `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Priority        string              `json:"priority"`
	ExpireTime      string              `json:"expireTime"`
	Scope           *Scope              `json:"scope"`
	FullPath        string              `json:"fullPath"`
	ApplicationName string              `json:"applicationName"`
	ApplicationID   uint                `json:"applicationID"`
	Tags            []controllertag.Tag `json:"tags"`

	// source info
	Git *codemodels.Git `json:"git"`

	// git config info
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
	Manifest       map[string]interface{}   `json:"manifest"`

	// status and update info
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedBy    *User     `json:"createdBy,omitempty"`
	UpdatedBy    *User     `json:"updatedBy,omitempty"`
	TTLInSeconds *uint     `json:"ttlInSeconds,omitempty"`
}

type WhetherLike struct {
	IsFavorite bool `json:"isFavorite"`
}
