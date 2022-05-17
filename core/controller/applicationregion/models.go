package applicationregion

import (
	"sort"

	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
)

type ApplicationRegion []*Region

func ofApplicationRegion(applicationRegions []*models.ApplicationRegion,
	environments []*envmodels.Environment, environmentRegions []*envregionmodels.EnvironmentRegion) ApplicationRegion {
	defaultRegionMap := make(map[string]string)
	for _, environmentRegion := range environmentRegions {
		defaultRegionMap[environmentRegion.EnvironmentName] = environmentRegion.RegionName
	}

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
			retApplicationRegions = append(retApplicationRegions, &Region{
				Environment: environment.Name,
				Region:      defaultRegionMap[environment.Name],
			})
		}
	}

	sort.Sort(RegionList(retApplicationRegions))
	return retApplicationRegions
}

type Region struct {
	Environment string `json:"environment"`
	Region      string `json:"region"`
}

type RegionList []*Region

func (r RegionList) Len() int {
	return len(r)
}

func (r RegionList) Less(i, j int) bool {
	const pre = "pre"
	const online = "online"
	if r[i].Environment == online {
		return false
	}
	if r[j].Environment == online {
		return true
	}
	if r[i].Environment == pre {
		return false
	}
	if r[j].Environment == pre {
		return true
	}
	return r[i].Environment < r[j].Environment
}

func (r RegionList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
