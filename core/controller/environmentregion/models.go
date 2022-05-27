package environmentregion

import (
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

type EnvironmentRegion struct {
	ID                uint   `json:"id"`
	Environment       string `json:"environment"`
	Region            string `json:"region"`
	RegionDisplayName string `json:"regionDisplayName"`
	IsDefault         bool   `json:"isDefault"`
	Disabled          bool   `json:"disabled"`
}

type EnvironmentRegions []*EnvironmentRegion

// ofEnvironmentModels []*models.Region to []*EnvironmentRegion
func ofRegionModels(regions []*regionmodels.Region,
	environmentRegions []*envregionmodels.EnvironmentRegion) EnvironmentRegions {
	displayNameMap := make(map[string]*regionmodels.Region)
	for _, region := range regions {
		displayNameMap[region.Name] = region
	}

	rs := make(EnvironmentRegions, 0)
	for _, envRegion := range environmentRegions {
		region := displayNameMap[envRegion.RegionName]
		rs = append(rs, &EnvironmentRegion{
			ID:                envRegion.ID,
			Region:            envRegion.RegionName,
			RegionDisplayName: region.DisplayName,
			Environment:       envRegion.EnvironmentName,
			IsDefault:         envRegion.IsDefault,
			Disabled:          region.Disabled,
		})
	}
	return rs
}

type CreateEnvironmentRegionRequest struct {
	EnvironmentName string `json:"environmentName"`
	RegionName      string `json:"regionName"`
}

type UpdateEnvironmentRegionRequest struct {
	IsDefault bool `json:"isDefault"`
	Disabled  bool `json:"disabled"`
}
