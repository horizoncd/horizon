package applicationregion

import (
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
)

type ApplicationRegion map[string]string

func ofApplicationRegion(applicationRegions []*models.ApplicationRegion,
	environments []*envmodels.Environment, regionConfig *region.Config) ApplicationRegion {
	retApplicationRegions := make(map[string]string)
	for _, applicationRegion := range applicationRegions {
		retApplicationRegions[applicationRegion.EnvironmentName] = applicationRegion.RegionName
	}

	// append default region
	for _, environment := range environments {
		if _, ok := retApplicationRegions[environment.Name]; !ok {
			defaultRegion := regionConfig.DefaultRegions[environment.Name]
			retApplicationRegions[environment.Name] = defaultRegion
		}
	}

	return retApplicationRegions
}
