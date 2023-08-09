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

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/admission"
	admissionmodels "github.com/horizoncd/horizon/pkg/admission/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
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

type CreateClusterParamsV2 struct {
	*CreateClusterRequestV2
	ApplicationID uint
	Environment   string
	Region        string
	// whether to merge json schema form data
	MergePatch bool
}

func (r *CreateClusterParamsV2) toAdmissionRequest(info *BuildTemplateInfo) *admission.Request {
	var gitURL, gitRef, gitSubfolder, gitRefType, image, templateName, templateRelease string
	if r.Git != nil {
		gitURL = r.Git.URL
		gitRef = r.Git.Ref()
		gitSubfolder = r.Git.Subfolder
		gitRefType = r.Git.RefType()
	}
	if r.Image != nil {
		image = *r.Image
	}
	if info.TemplateInfo != nil {
		templateName = info.TemplateInfo.Name
		templateRelease = info.TemplateInfo.Release
	}
	cluster := &Clusterv2{
		Cluster: &models.Cluster{
			ApplicationID:   r.ApplicationID,
			Name:            r.Name,
			EnvironmentName: r.Environment,
			RegionName:      r.Region,
			Description:     r.Description,
			GitURL:          gitURL,
			GitSubfolder:    gitSubfolder,
			GitRef:          gitRef,
			GitRefType:      gitRefType,
			Image:           image,
			Template:        templateName,
			TemplateRelease: templateRelease,
		},
		TemplateConfig: info.TemplateConfig,
		BuildConfig:    info.BuildConfig,
		Tags:           r.Tags,
	}
	return &admission.Request{
		Operation: admissionmodels.OperationCreate,
		Resource:  common.ResourceCluster,
		Version:   common.Version1,
		Object:    cluster,
		OldObject: nil,
		Options:   map[string]interface{}{},
	}
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

func clusterV2FromAdmission(req *admission.Request) (*Clusterv2, []*tagmodels.Tag, error) {
	admissionCluster, ok := req.Object.(*Clusterv2)
	if !ok {
		return nil, nil, perror.Wrap(herrors.ErrParamInvalid, "could not convert admission request to cluster")
	}
	if admissionCluster.Tags == nil {
		return admissionCluster, nil, nil
	}
	tags := make([]*tagmodels.Tag, 0)
	for _, tag := range admissionCluster.Tags {
		tags = append(tags, &tagmodels.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return admissionCluster, tags, nil
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

func (r *UpdateClusterRequestV2) toAdmissionRequest(
	oldCluster *models.Cluster,
) *admission.Request {
	var gitURL, gitSubFolder, gitRef, gitRefType,
		image, region, env, templateName, templateRelease string

	if r.Git != nil {
		gitURL, gitSubFolder, gitRefType, gitRef = r.Git.URL,
			r.Git.Subfolder, r.Git.RefType(), r.Git.Ref()
	} else {
		gitURL = oldCluster.GitURL
		gitSubFolder = oldCluster.GitSubfolder
		gitRefType = oldCluster.GitRefType
		gitRef = oldCluster.GitRef
	}
	if r.Image != nil {
		image = *r.Image
	}
	if r.Region != nil {
		region = *r.Region
	} else {
		region = oldCluster.RegionName
	}
	if r.Environment != nil {
		env = *r.Environment
	} else {
		env = oldCluster.EnvironmentName
	}

	if r.TemplateInfo != nil {
		templateName = r.TemplateInfo.Name
		templateRelease = r.TemplateInfo.Release
	} else {
		templateName = oldCluster.Template
		templateRelease = oldCluster.TemplateRelease
	}

	cluster := &Clusterv2{
		Cluster: &models.Cluster{
			ApplicationID:   oldCluster.ApplicationID,
			EnvironmentName: env,
			RegionName:      region,
			Description:     r.Description,
			GitURL:          gitURL,
			GitSubfolder:    gitSubFolder,
			GitRef:          gitRef,
			GitRefType:      gitRefType,
			Image:           image,
			Template:        templateName,
			TemplateRelease: templateRelease,
		},
		TemplateInfo:   &codemodels.TemplateInfo{Name: templateName, Release: templateRelease},
		TemplateConfig: r.TemplateConfig,
		BuildConfig:    r.BuildConfig,
		Tags:           r.Tags,
	}

	return &admission.Request{
		Operation: admissionmodels.OperationUpdate,
		Resource:  common.ResourceCluster,
		Version:   common.Version1,
		Object:    cluster,
		OldObject: &Clusterv2{
			Cluster: oldCluster,
		},
		Options: map[string]interface{}{},
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
