// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package environmentregion

import (
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
)

type EnvironmentRegion struct {
	ID                uint   `json:"id"`
	EnvironmentName   string `json:"environmentName"`
	RegionName        string `json:"regionName"`
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
			RegionName:        envRegion.RegionName,
			RegionDisplayName: region.DisplayName,
			EnvironmentName:   envRegion.EnvironmentName,
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
