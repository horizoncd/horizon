package cluster

import (
	"time"

	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
)

type Base struct {
	Description   string         `json:"description"`
	Git           *Git           `json:"git"`
	TemplateInput *TemplateInput `json:"templateInput"`
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

	Name string `json:"name"`
}

type UpdateClusterRequest struct {
	*Base

	Template *Template `json:"template"`
}

type GetClusterResponse struct {
	*CreateClusterRequest

	ID          uint         `json:"id"`
	FullPath    string       `json:"fullPath"`
	Application *Application `json:"application"`
	Priority    string       `json:"priority"`
	Template    *Template    `json:"template"`
	Scope       *Scope       `json:"scope"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
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
	Environment string `json:"environment"`
	Region      string `json:"region"`
}

func (r *CreateClusterRequest) toClusterModel(application *appmodels.Application,
	er *envmodels.EnvironmentRegion) *models.Cluster {
	return &models.Cluster{
		ApplicationID:       application.ID,
		Name:                r.Name,
		Description:         r.Description,
		GitURL:              application.GitURL,
		GitSubfolder:        application.GitSubfolder,
		GitBranch:           r.Git.Branch,
		Template:            application.Template,
		TemplateRelease:     application.TemplateRelease,
		EnvironmentRegionID: er.ID,
	}
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
	}
}

func ofClusterModel(application *appmodels.Application, cluster *models.Cluster,
	er *envmodels.EnvironmentRegion, fullPath string,
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
			Name: cluster.Name,
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
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

type ListClusterResponse struct {
	ID       uint      `json:"id"`
	Name     string    `json:"name"`
	Scope    *Scope    `json:"scope"`
	Template *Template `json:"template"`
}

func ofClustersWithEnvAndRegion(clusters []*models.ClusterWithEnvAndRegion) []*ListClusterResponse {
	respList := make([]*ListClusterResponse, 0)
	for _, c := range clusters {
		respList = append(respList, &ListClusterResponse{
			ID:   c.ID,
			Name: c.Name,
			Scope: &Scope{
				Environment: c.EnvironmentName,
				Region:      c.RegionName,
			},
			Template: &Template{
				Name:    c.Template,
				Release: c.TemplateRelease,
			},
		})
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
}
