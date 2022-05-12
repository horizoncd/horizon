package environment

import (
	"g.hz.netease.com/horizon/pkg/environment/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

type Environment struct {
	Name                     string `json:"name"`
	DisplayName              string `json:"displayName"`
	DefaultRegion            string `json:"defaultRegion"`
	DefaultRegionDisplayName string `json:"defaultRegionDisplayName"`
}

type Environments []*Environment

// ofEnvironmentModels []*models.Environment to []*Environment
func ofEnvironmentModels(envs []*models.Environment) Environments {
	environments := make(Environments, 0)
	for _, env := range envs {
		environments = append(environments, &Environment{
			Name:          env.Name,
			DisplayName:   env.DisplayName,
			DefaultRegion: env.DefaultRegion,
		})
	}
	return environments
}

type Region struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type Regions []*Region

// ofEnvironmentModels []*models.Region to []*Region
func ofRegionModels(regions []*regionmodels.Region) Regions {
	rs := make(Regions, 0)
	for _, region := range regions {
		rs = append(rs, &Region{
			Name:        region.Name,
			DisplayName: region.DisplayName,
		})
	}
	return rs
}

type UpdateEnvironmentRequest struct {
	DisplayName   string `json:"displayName"`
	DefaultRegion string `json:"defaultRegion"`
}
