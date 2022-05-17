package environment

import (
	"g.hz.netease.com/horizon/pkg/environment/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

type Environment struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type Environments []*Environment

// ofEnvironmentModels []*models.Environment to []*Environment
func ofEnvironmentModels(envs []*models.Environment) Environments {
	environments := make(Environments, 0)
	for _, env := range envs {
		environments = append(environments, &Environment{
			Name:        env.Name,
			DisplayName: env.DisplayName,
		})
	}
	return environments
}

type CreateEnvironmentRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type UpdateEnvironmentRequest struct {
	DisplayName string `json:"displayName"`
}

type Region struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type Regions []*Region

// ofEnvironmentModels []*models.Region to []*Region
func ofRegionModels(regions []*regionmodels.Region, environmentRegions []*envregionmodels.EnvironmentRegion) Regions {
	displayNameMap := make(map[string]string)
	for _, region := range regions {
		displayNameMap[region.Name] = region.DisplayName
	}

	rs := make(Regions, 0)
	for _, region := range environmentRegions {
		rs = append(rs, &Region{
			Name:        region.RegionName,
			DisplayName: displayNameMap[region.RegionName],
		})
	}
	return rs
}
