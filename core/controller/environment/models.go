package environment

import (
	"g.hz.netease.com/horizon/pkg/environment/models"
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

type CreateEnvironmentRequest struct {
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	DefaultRegion string `json:"defaultRegion"`
}

type UpdateEnvironmentRequest struct {
	DisplayName   string `json:"displayName"`
	DefaultRegion string `json:"defaultRegion"`
}
