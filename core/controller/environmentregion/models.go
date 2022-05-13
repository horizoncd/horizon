package environmentregion

import (
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

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

type CreateEnvironmentRegionRequest struct {
	EnvironmentName string `json:"environmentName"`
	RegionName      string `json:"regionName"`
}

type UpdateEnvironmentRegionRequest struct {
	EnvironmentName string `json:"environmentName"`
	RegionName      string `json:"regionName"`
}
