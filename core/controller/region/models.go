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

package region

import (
	"time"

	"github.com/horizoncd/horizon/core/controller/registry"
	"github.com/horizoncd/horizon/core/controller/tag"
	"github.com/horizoncd/horizon/pkg/models"
)

type Region struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	DisplayName   string            `json:"displayName"`
	Server        string            `json:"server"`
	Certificate   string            `json:"certificate"`
	IngressDomain string            `json:"ingressDomain"`
	PrometheusURL string            `json:"prometheusURL"`
	Disabled      bool              `json:"disabled"`
	RegistryID    uint              `json:"registryID"`
	Registry      registry.Registry `json:"registry"`
	Tags          []tag.Tag         `json:"tags"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

type CreateRegionRequest struct {
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	Server        string `json:"server"`
	Certificate   string `json:"certificate"`
	IngressDomain string `json:"ingressDomain"`
	PrometheusURL string `json:"prometheusURL"`
	RegistryID    uint   `json:"registryID"`
}

type UpdateRegionRequest struct {
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	Server        string `json:"server"`
	Certificate   string `json:"certificate"`
	IngressDomain string `json:"ingressDomain"`
	PrometheusURL string `json:"prometheusURL"`
	RegistryID    uint   `json:"registryID"`
	Disabled      bool   `json:"disabled"`
}

func ofRegionEntity(entity *models.RegionEntity) *Region {
	var tags []tag.Tag
	for _, t := range entity.Tags {
		tags = append(tags, tag.Tag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	r := &Region{
		ID:            entity.ID,
		Name:          entity.Name,
		DisplayName:   entity.DisplayName,
		Server:        entity.Server,
		IngressDomain: entity.IngressDomain,
		PrometheusURL: entity.PrometheusURL,
		Certificate:   entity.Certificate,
		Disabled:      entity.Disabled,
		RegistryID:    entity.RegistryID,
		Tags:          tags,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
	if entity.Registry != nil {
		r.Registry = registry.Registry{
			Name: entity.Registry.Name,
		}
	}
	return r
}

func ofRegionEntities(entities []*models.RegionEntity) []*Region {
	regions := make([]*Region, 0)

	for _, entity := range entities {
		regions = append(regions, ofRegionEntity(entity))
	}

	return regions
}
