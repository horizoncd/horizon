package cluster

import (
	"time"

	"g.hz.netease.com/horizon/core/common"
	controllertag "g.hz.netease.com/horizon/core/controller/tag"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

type Base struct {
	Description   string               `json:"description"`
	Git           *Git                 `json:"git"`
	Template      *Template            `json:"template"`
	TemplateInput *TemplateInput       `json:"templateInput"`
	Tags          []*controllertag.Tag `json:"tags"`
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

type CreateClusterRequest struct {
	*Base

	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// TODO(gjq): remove these two params after migration
	Image        string            `json:"image"`
	ExtraMembers map[string]string `json:"extraMembers"`
}

type UpdateClusterRequest struct {
	*Base
	Environment string `json:"environment"`
	Region      string `json:"region"`
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
	er *envregionmodels.EnvironmentRegion) (*models.Cluster, []*tagmodels.Tag) {
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
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitBranch:       r.Git.Branch,
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
	var gitURL, gitSubfolder, gitBranch string
	if r.Git == nil || r.Git.URL == "" {
		gitURL = cluster.GitURL
	} else {
		gitURL = r.Git.URL
	}
	gitSubfolder = r.Git.Subfolder
	if r.Git == nil || r.Git.Branch == "" {
		gitBranch = cluster.GitBranch
	} else {
		gitBranch = r.Git.Branch
	}
	return &models.Cluster{
		EnvironmentName: er.EnvironmentName,
		RegionName:      er.RegionName,
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

func ofClusterModel(application *appmodels.Application, cluster *models.Cluster, fullPath, namespace string,
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
			Environment: cluster.EnvironmentName,
			Region:      cluster.RegionName,
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
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Scope       *Scope                `json:"scope"`
	Template    *Template             `json:"template"`
	Git         *GitResponse          `json:"git"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
	Tags        []*tagmodels.TagBasic `json:"tags,omitempty"`
}

func ofClusterWithEnvAndRegion(cluster *models.ClusterWithRegion) *ListClusterResponse {
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
