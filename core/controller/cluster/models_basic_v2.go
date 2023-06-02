// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"time"

	"github.com/horizoncd/horizon/core/common"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
)

type CreateClusterRequestV2 struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Priority    string              `json:"priority"`
	ExpireTime  string              `json:"expireTime"`
	Git         *codemodels.Git     `json:"git"`
	Image       *string             `json:"image"`
	Tags        tagmodels.TagsBasic `json:"tags"`

	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`

	// TODO(tom): just for internal usage
	ExtraMembers map[string]string `json:"extraMembers"`
}

func (r *CreateClusterRequestV2) toClusterModel(application *appmodels.Application,
	er *envregionmodels.EnvironmentRegion, info *BuildTemplateInfo,
	expireSeconds uint) (*models.Cluster, []*tagmodels.Tag) {
	cluster := &models.Cluster{
		ApplicationID:   application.ID,
		Name:            r.Name,
		EnvironmentName: er.EnvironmentName,
		RegionName:      er.RegionName,
		Description:     r.Description,
		ExpireSeconds:   expireSeconds,
		Template:        info.TemplateInfo.Name,
		TemplateRelease: info.TemplateInfo.Release,
		Status:          common.ClusterStatusCreating,
	}
	cluster.GitURL = func() string {
		if r.Git == nil {
			return application.GitURL
		}
		if r.Git.URL == "" && application.GitURL != "" {
			return application.GitURL
		}
		// if URL is empty string, this means this cluster not depends on build from git
		return r.Git.URL
	}()
	cluster.GitSubfolder = func() string {
		if r.Git == nil || r.Git.Subfolder == "" {
			return application.GitSubfolder
		}
		return r.Git.Subfolder
	}()
	cluster.GitRef = func() string {
		if r.Git == nil {
			return application.GitRef
		}
		return r.Git.Ref()
	}()
	cluster.GitRefType = func() string {
		if r.Git == nil {
			return application.GitRefType
		}
		return r.Git.RefType()
	}()
	cluster.Image = func() string {
		if r.Image == nil {
			return application.Image
		}
		return *r.Image
	}()
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

	Tags tagmodels.TagsBasic `json:"tags"`
	// source info
	Git   *codemodels.Git `json:"git"`
	Image *string         `json:"image"`

	// git config info
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
}

func (r *UpdateClusterRequestV2) toClusterModel(cluster *models.Cluster, expireSeconds uint, environmentName,
	regionName, templateName, templateRelease string) (*models.Cluster, []*tagmodels.Tag) {
	var gitURL, gitSubFolder, gitRef, gitRefType, image string
	if r.Git != nil {
		gitURL, gitSubFolder, gitRefType, gitRef = r.Git.URL,
			r.Git.Subfolder, r.Git.RefType(), r.Git.Ref()
	} else {
		gitURL = cluster.GitURL
		gitSubFolder = cluster.GitSubfolder
		gitRefType = cluster.GitRefType
		gitRef = cluster.GitRef
	}
	if r.Image != nil {
		image = *r.Image
	} else {
		image = cluster.Image
	}

	tags := make([]*tagmodels.Tag, 0)
	for _, tag := range r.Tags {
		tags = append(tags, &tagmodels.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
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
		Image:           image,
		Template:        templateName,
		TemplateRelease: templateRelease,
	}, tags
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
	Tags            tagmodels.TagsBasic `json:"tags"`

	// source info
	Git   *codemodels.Git `json:"git"`
	Image string          `json:"image"`

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

type CreateClusterParamsV2 struct {
	*CreateClusterRequestV2
	ApplicationID uint
	Environment   string
	Region        string
	// whether to merge json schema form data
	MergePatch bool
}
