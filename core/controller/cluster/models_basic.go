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
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/admission"
	admissionmodels "github.com/horizoncd/horizon/pkg/admission/models"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

type Base struct {
	Description   string                `json:"description"`
	Git           *codemodels.Git       `json:"git"`
	Template      *Template             `json:"template"`
	TemplateInput *TemplateInput        `json:"templateInput"`
	Tags          []*tagmodels.TagBasic `json:"tags"`
}

type TemplateInput struct {
	Application map[string]interface{} `json:"application"`
	Pipeline    map[string]interface{} `json:"pipeline"`
}

type CreateClusterRequest struct {
	*Base

	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	ExpireTime string `json:"expireTime"`
	// TODO(gjq): remove these two params after migration
	Image        string            `json:"image"`
	ExtraMembers map[string]string `json:"extraMembers"`
}

type UpdateClusterRequest struct {
	*Base
	Environment string `json:"environment"`
	Region      string `json:"region"`
	ExpireTime  string `json:"expireTime"`
}

type GetClusterResponse struct {
	*CreateClusterRequest

	ID                   uint         `json:"id"`
	FullPath             string       `json:"fullPath"`
	Application          *Application `json:"application"`
	Priority             string       `json:"priority"`
	Template             *Template    `json:"template"`
	Scope                *Scope       `json:"scope"`
	LatestDeployedCommit string       `json:"latestDeployedCommit,omitempty"`
	Status               string       `json:"status,omitempty"`
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
	TTLInSeconds         *uint        `json:"ttlInSeconds"`
	CreatedBy            *User        `json:"createdBy,omitempty"`
	UpdatedBy            *User        `json:"updatedBy,omitempty"`
}

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Application struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Template struct {
	Name    string `json:"name"`
	Release string `json:"release"`
}

type Scope struct {
	Environment       string `json:"environment"`
	Region            string `json:"region"`
	RegionDisplayName string `json:"regionDisplayName,omitempty"`
}

func (r *CreateClusterRequest) toClusterModel(application *appmodels.Application,
	env, region string) *models.ClusterWithTags {
	var (
		// r.Git cannot be nil
		gitURL          = r.Git.URL
		gitSubfolder    = r.Git.Subfolder
		templateName    = ""
		templateRelease = ""
	)
	// if gitURL or gitSubfolder is empty, use application's gitURL or gitSubfolder
	if gitURL == "" {
		gitURL = application.GitURL
	}
	if gitSubfolder == "" {
		gitSubfolder = application.GitSubfolder
	}
	if r.Template != nil {
		templateName = r.Template.Name
		templateRelease = r.Template.Release
	}
	cluster := &models.Cluster{
		ApplicationID:   application.ID,
		Name:            r.Name,
		EnvironmentName: env,
		RegionName:      region,
		Description:     r.Description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          r.Git.Ref(),
		GitRefType:      r.Git.RefType(),
		Template:        templateName,
		TemplateRelease: templateRelease,
	}
	tags := make([]*tagmodels.Tag, 0)
	for _, tag := range r.Tags {
		tags = append(tags, &tagmodels.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return &models.ClusterWithTags{
		Cluster: cluster,
		Tags:    tags,
	}
}

func (r *CreateClusterRequest) toAdmissionRequest(
	app *appmodels.Application,
	env, region string,
	mergePatch bool) *admission.Request {
	if r.Template == nil {
		r.Template = &Template{
			Name:    app.Template,
			Release: app.TemplateRelease,
		}
	}
	cluster := r.toClusterModel(app, env, region)
	admissionCluster := &Cluster{
		Cluster:       cluster.Cluster,
		TemplateInput: r.TemplateInput,
		Tags:          r.Tags,
	}

	return &admission.Request{
		Operation: admissionmodels.OperationCreate,
		Resource:  common.ResourceCluster,
		Version:   common.Version1,
		Object:    admissionCluster,
		OldObject: nil,
		Options: map[string]interface{}{
			"mergePatch": mergePatch,
		},
	}
}

func clusterFromAdmission(req *admission.Request) (*Cluster, []*tagmodels.Tag, error) {
	admissionCluster, ok := req.Object.(*Cluster)
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

func (r *UpdateClusterRequest) toClusterModel(cluster *models.Cluster) (*models.Cluster, []*tagmodels.Tag) {
	var gitURL, gitSubfolder, gitRef, gitRefType string
	if r.Git != nil {
		gitURL, gitSubfolder, gitRefType, gitRef = r.Git.URL,
			r.Git.Subfolder, r.Git.RefType(), r.Git.Ref()
	} else {
		gitURL = cluster.GitURL
		gitSubfolder = cluster.GitSubfolder
		gitRefType = cluster.GitRefType
		gitRef = cluster.GitRef
	}

	tags := make([]*tagmodels.Tag, 0)
	if r.Tags == nil {
		tags = nil
	}
	for _, tag := range r.Tags {
		tags = append(tags, &tagmodels.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	var templateRelease string
	if r.Template == nil || r.Template.Release == "" {
		templateRelease = cluster.TemplateRelease
	} else {
		templateRelease = r.Template.Release
	}

	env := cluster.EnvironmentName
	region := cluster.RegionName
	if r.Environment != "" {
		env = r.Environment
	}
	if r.Region != "" {
		region = r.Region
	}

	return &models.Cluster{
		ApplicationID:   cluster.ApplicationID,
		Name:            cluster.Name,
		EnvironmentName: env,
		RegionName:      region,
		Description:     r.Description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitRef,
		GitRefType:      gitRefType,
		Template:        cluster.Template,
		TemplateRelease: templateRelease,
		Status:          cluster.Status,
		ExpireSeconds:   cluster.ExpireSeconds,
	}, tags
}

func (r *UpdateClusterRequest) toAdmissionRequest(
	oldCluster *models.Cluster,
	oldTags []*tagmodels.Tag,
	mergePatch bool,
) *admission.Request {
	newCluster, _ := r.toClusterModel(oldCluster)
	newAdmissionCluster := &Cluster{
		Cluster:       newCluster,
		TemplateInput: r.TemplateInput,
		Tags:          r.Tags,
	}

	oldBasicTag := make([]*tagmodels.TagBasic, 0)
	for _, tag := range oldTags {
		oldBasicTag = append(oldBasicTag, &tagmodels.TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	oldAdmCluster := &Cluster{
		Cluster: oldCluster,
		Tags:    oldBasicTag,
	}

	return &admission.Request{
		Operation: admissionmodels.OperationUpdate,
		Resource:  common.ResourceCluster,
		Version:   common.Version1,
		Object:    newAdmissionCluster,
		OldObject: oldAdmCluster,
		Options: map[string]interface{}{
			"mergePatch": mergePatch,
		},
	}
}

func getUserFromMap(id uint, userMap map[uint]*usermodels.User) *usermodels.User {
	user, ok := userMap[id]
	if !ok {
		return nil
	}
	return user
}

func toUser(user *usermodels.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		ID:    user.ID,
		Name:  user.FullName,
		Email: user.Email,
	}
}

func ofClusterModel(application *appmodels.Application, cluster *models.Cluster, fullPath, namespace string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}, tags ...*tagmodels.Tag) *GetClusterResponse {
	expireTime := ""
	if cluster.ExpireSeconds > 0 {
		expireTime = time.Duration(cluster.ExpireSeconds * 1e9).String()
	}

	return &GetClusterResponse{
		CreateClusterRequest: &CreateClusterRequest{
			Base: &Base{
				Description: cluster.Description,
				Tags:        tagmodels.Tags(tags).IntoTagsBasic(),
				Git: codemodels.NewGit(cluster.GitURL, cluster.GitSubfolder,
					cluster.GitRefType, cluster.GitRef),
				TemplateInput: &TemplateInput{
					Application: applicationJSONBlob,
					Pipeline:    pipelineJSONBlob,
				},
			},
			Name:       cluster.Name,
			Namespace:  namespace,
			ExpireTime: expireTime,
		},
		ID:       cluster.ID,
		FullPath: fullPath,
		Application: &Application{
			ID:   application.ID,
			Name: application.Name,
		},
		Priority: string(application.Priority),
		Template: &Template{
			Name:    cluster.Template,
			Release: cluster.TemplateRelease,
		},
		Scope: &Scope{
			Environment: cluster.EnvironmentName,
			Region:      cluster.RegionName,
		},
		Status:    cluster.Status,
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

type GitResponse struct {
	GitURL  string `json:"gitURL"`
	HTTPURL string `json:"httpURL"`
}

type ListClusterResponse struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Type        *string               `json:"type,omitempty"`
	Description string                `json:"description"`
	Scope       *Scope                `json:"scope"`
	Template    *Template             `json:"template"`
	Git         *GitResponse          `json:"git"`
	IsFavorite  *bool                 `json:"isFavorite"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
	Tags        []*tagmodels.TagBasic `json:"tags,omitempty"`
}

func ofCluster(cluster *models.Cluster) *ListClusterResponse {
	return &ListClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		Type:        cluster.Type,
		Scope: &Scope{
			Environment: cluster.EnvironmentName,
			Region:      cluster.RegionName,
		},
		Template: &Template{
			Name:    cluster.Template,
			Release: cluster.TemplateRelease,
		},
		Git: &GitResponse{
			GitURL: cluster.GitURL,
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

func ofClusterWithEnvAndRegion(cluster *models.ClusterWithRegion) *ListClusterResponse {
	resp := ofCluster(cluster.Cluster)
	resp.Scope.RegionDisplayName = cluster.RegionDisplayName
	return resp
}

func ofClustersWithEnvRegionTags(clusters []*models.ClusterWithRegion, tags []*tagmodels.Tag) []*ListClusterResponse {
	tagMap := map[uint][]*tagmodels.TagBasic{}
	for _, tag := range tags {
		tagBasic := &tagmodels.TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		}
		tagMap[tag.ResourceID] = append(tagMap[tag.ResourceID], tagBasic)
	}

	respList := make([]*ListClusterResponse, 0)
	for _, c := range clusters {
		cluster := ofClusterWithEnvAndRegion(c)
		cluster.Tags = tagMap[c.ID]
		respList = append(respList, cluster)
	}
	return respList
}

type GetClusterByNameResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Template    *Template       `json:"template"`
	Git         *codemodels.Git `json:"git"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	FullPath    string          `json:"fullPath"`
}

type ListClusterWithFullResponse struct {
	*ListClusterResponse
	IsFavorite *bool  `json:"isFavorite,omitempty"`
	FullName   string `json:"fullName,omitempty"`
	FullPath   string `json:"fullPath,omitempty"`
}

type ListClusterWithExpiryResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	EnvironmentName string    `json:"environmentName"`
	Status          string    `json:"status"`
	ExpireSeconds   uint      `json:"expireSeconds"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func ofClusterWithExpiry(clusters []*models.Cluster) []*ListClusterWithExpiryResponse {
	resList := make([]*ListClusterWithExpiryResponse, 0, len(clusters))
	for _, c := range clusters {
		resList = append(resList, &ListClusterWithExpiryResponse{
			ID:              c.ID,
			Name:            c.Name,
			EnvironmentName: c.EnvironmentName,
			Status:          c.Status,
			ExpireSeconds:   c.ExpireSeconds,
			UpdatedAt:       c.UpdatedAt,
		})
	}
	return resList
}
