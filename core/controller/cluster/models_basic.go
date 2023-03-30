package cluster

import (
	"time"

	controllertag "github.com/horizoncd/horizon/core/controller/tag"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

type Base struct {
	Description   string               `json:"description"`
	Git           *codemodels.Git      `json:"git"`
	Template      *Template            `json:"template"`
	TemplateInput *TemplateInput       `json:"templateInput"`
	Tags          []*controllertag.Tag `json:"tags"`
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
	er *envregionmodels.EnvironmentRegion, expireSeconds uint) (*models.Cluster, []*tagmodels.Tag) {
	var (
		// r.Git cannot be nil
		gitURL       = r.Git.URL
		gitSubfolder = r.Git.Subfolder
	)
	// if gitURL or gitSubfolder is empty, use application's gitURL or gitSubfolder
	if gitURL == "" {
		gitURL = application.GitURL
	}
	if gitSubfolder == "" {
		gitSubfolder = application.GitSubfolder
	}
	cluster := &models.Cluster{
		ApplicationID:   application.ID,
		Name:            r.Name,
		EnvironmentName: er.EnvironmentName,
		RegionName:      er.RegionName,
		Description:     r.Description,
		ExpireSeconds:   expireSeconds,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          r.Git.Ref(),
		GitRefType:      r.Git.RefType(),
		Template:        r.Template.Name,
		TemplateRelease: r.Template.Release,
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

func (r *UpdateClusterRequest) toClusterModel(cluster *models.Cluster,
	templateRelease string, er *envregionmodels.EnvironmentRegion) *models.Cluster {
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

	return &models.Cluster{
		EnvironmentName: er.EnvironmentName,
		RegionName:      er.RegionName,
		Description:     r.Description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitRef,
		GitRefType:      gitRefType,
		Template:        cluster.Template,
		TemplateRelease: templateRelease,
		Status:          cluster.Status,
		ExpireSeconds:   cluster.ExpireSeconds,
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
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) *GetClusterResponse {
	expireTime := ""
	if cluster.ExpireSeconds > 0 {
		expireTime = time.Duration(cluster.ExpireSeconds * 1e9).String()
	}
	return &GetClusterResponse{
		CreateClusterRequest: &CreateClusterRequest{
			Base: &Base{
				Description: cluster.Description,
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
		IsFavorite: cluster.IsFavorite,
		CreatedAt:  cluster.CreatedAt,
		UpdatedAt:  cluster.UpdatedAt,
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
	FullName string `json:"fullName,omitempty"`
	FullPath string `json:"fullPath,omitempty"`
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
