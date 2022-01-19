package cluster

import (
	"time"

	"g.hz.netease.com/horizon/core/common"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

type Base struct {
	Description   string         `json:"description"`
	Git           *Git           `json:"git"`
	Template      *Template      `json:"template"`
	TemplateInput *TemplateInput `json:"templateInput"`
	Tags          []*Tag         `json:"tags"`
}

type TemplateInput struct {
	Application map[string]interface{} `json:"application"`
	Pipeline    map[string]interface{} `json:"pipeline"`
}

type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateClusterRequest struct {
	*Base

	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// TODO(gjq): remove these two params after migration
	Image string `json:"image"`
}

type UpdateClusterRequest struct {
	*Base
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
	er *envmodels.EnvironmentRegion) (*models.Cluster, []*clustertagmodels.ClusterTag) {
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
		ApplicationID:       application.ID,
		Name:                r.Name,
		Description:         r.Description,
		GitURL:              gitURL,
		GitSubfolder:        gitSubfolder,
		GitBranch:           r.Git.Branch,
		Template:            r.Template.Name,
		TemplateRelease:     r.Template.Release,
		EnvironmentRegionID: er.ID,
	}
	clusterTags := make([]*clustertagmodels.ClusterTag, 0)
	for _, tag := range r.Tags {
		clusterTags = append(clusterTags, &clustertagmodels.ClusterTag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return cluster, clusterTags
}

func (r *UpdateClusterRequest) toClusterModel(cluster *models.Cluster,
	templateRelease string) *models.Cluster {
	var gitURL, gitSubfolder, gitBranch string
	if r.Git == nil || r.Git.URL == "" {
		gitURL = cluster.GitURL
	} else {
		gitURL = r.Git.URL
	}
	if r.Git == nil || r.Git.Subfolder == "" {
		gitSubfolder = cluster.GitSubfolder
	} else {
		gitSubfolder = r.Git.Subfolder
	}
	if r.Git == nil || r.Git.Branch == "" {
		gitBranch = cluster.GitBranch
	} else {
		gitBranch = r.Git.Branch
	}
	return &models.Cluster{
		Description:     r.Description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitBranch:       gitBranch,
		TemplateRelease: templateRelease,
		Status:          cluster.Status,
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

func ofClusterModel(application *appmodels.Application, cluster *models.Cluster,
	er *envmodels.EnvironmentRegion, fullPath, namespace string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) *GetClusterResponse {
	return &GetClusterResponse{
		CreateClusterRequest: &CreateClusterRequest{
			Base: &Base{
				Description: cluster.Description,
				Git: &Git{
					URL:       cluster.GitURL,
					Subfolder: cluster.GitSubfolder,
					Branch:    cluster.GitBranch,
				},
				TemplateInput: &TemplateInput{
					Application: applicationJSONBlob,
					Pipeline:    pipelineJSONBlob,
				},
			},
			Name:      cluster.Name,
			Namespace: namespace,
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
			Environment: er.EnvironmentName,
			Region:      er.RegionName,
		},
		Status:    cluster.Status,
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

type GitResponse struct {
	SSHURL  string `json:"sshURL"`
	HTTPURL string `json:"httpURL"`
}

type ListClusterResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Scope       *Scope       `json:"scope"`
	Template    *Template    `json:"template"`
	Git         *GitResponse `json:"git"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

func ofClusterWithEnvAndRegion(cluster *models.ClusterWithEnvAndRegion) *ListClusterResponse {
	return &ListClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		Scope: &Scope{
			Environment:       cluster.EnvironmentName,
			Region:            cluster.RegionName,
			RegionDisplayName: cluster.RegionDisplayName,
		},
		Template: &Template{
			Name:    cluster.Template,
			Release: cluster.TemplateRelease,
		},
		Git: &GitResponse{
			SSHURL:  cluster.GitURL,
			HTTPURL: common.InternalSSHToHTTPURL(cluster.GitURL),
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

func ofClustersWithEnvAndRegion(clusters []*models.ClusterWithEnvAndRegion) []*ListClusterResponse {
	respList := make([]*ListClusterResponse, 0)
	for _, c := range clusters {
		respList = append(respList, ofClusterWithEnvAndRegion(c))
	}
	return respList
}

type GetClusterByNameResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Template    *Template `json:"template"`
	Git         *Git      `json:"git"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	FullPath    string    `json:"fullPath"`
}

type ListClusterWithFullResponse struct {
	*ListClusterResponse
	FullName string `json:"fullName"`
	FullPath string `json:"fullPath"`
}
