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

package models

import (
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
)

type Region struct {
	global.Model

	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	PrometheusURL string
	RegistryID    uint `gorm:"column:registry_id"`
	Disabled      bool
	CreatedBy     uint
	UpdatedBy     uint
}

// RegionEntity region entity, region with registry
type RegionEntity struct {
	*Region

	Registry *registrymodels.Registry
	Tags     []*tagmodels.Tag
}

type RegionPart struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Disabled    bool   `json:"disabled"`
	IsDefault   bool   `json:"isDefault"`
}

type RegionParts []*RegionPart
