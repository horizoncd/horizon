package applicationregion

import (
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
)

type ApplicationRegion []*Region

func ofApplicationRegion(applicationRegions []*models.ApplicationRegion,
	environments []*envmodels.Environment, regionConfig *region.Config) ApplicationRegion {
	retApplicationRegions := make([]*Region, 0)
	envMap := make(map[string]bool)
	for _, applicationRegion := range applicationRegions {
		retApplicationRegions = append(retApplicationRegions, &Region{
			Environment: applicationRegion.EnvironmentName,
			Region:      applicationRegion.RegionName,
		})
		envMap[applicationRegion.EnvironmentName] = true
	}

	// append default region
	for _, environment := range environments {
		if _, ok := envMap[environment.Name]; !ok {
			defaultRegion := regionConfig.DefaultRegions[environment.Name]
			retApplicationRegions = append(retApplicationRegions, &Region{
				Environment: environment.Name,
				Region:      defaultRegion,
			})
		}
	}

	return retApplicationRegions
}

type Region struct {
	Environment string `json:"environment"`
	Region      string `json:"region"`
}
