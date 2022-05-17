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
	displayNameMap := make(map[string]string)
	for _, region := range regions {
		displayNameMap[region.Name] = region.DisplayName
	}

	rs := make(EnvironmentRegions, 0)
	for _, region := range environmentRegions {
		rs = append(rs, &EnvironmentRegion{
			ID:                region.ID,
			Region:            region.RegionName,
			RegionDisplayName: displayNameMap[region.RegionName],
			Environment:       region.EnvironmentName,
			IsDefault:         region.IsDefault,
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
