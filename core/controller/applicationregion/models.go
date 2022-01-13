package applicationregion

import (
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
)

type ApplicationRegion map[string]*Region

func ofApplicationRegion(applicationRegions []*models.ApplicationRegion, regions []*regionmodels.Region,
	environments []*envmodels.Environment, regionConfig *region.Config) ApplicationRegion {
	regionMap := make(map[string]string)
	for _, r := range regions {
		regionMap[r.Name] = r.DisplayName
	}

	retApplicationRegions := make(map[string]*Region)
	for _, applicationRegion := range applicationRegions {
		retApplicationRegions[applicationRegion.EnvironmentName] = &Region{
			Region:            applicationRegion.RegionName,
			RegionDisplayName: regionMap[applicationRegion.RegionName],
		}
	}

	// append default region
	for _, environment := range environments {
		if _, ok := retApplicationRegions[environment.Name]; !ok {
			defaultRegion := regionConfig.DefaultRegions[environment.Name]
			retApplicationRegions[environment.Name] = &Region{
				Region:            defaultRegion,
				RegionDisplayName: regionMap[defaultRegion],
			}
		}
	}

	return retApplicationRegions
}

type Region struct {
	Region            string `json:"region"`
	RegionDisplayName string `json:"regionDisplayName"`
}
