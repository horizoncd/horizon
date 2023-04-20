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

package registry

import (
	"time"

	"github.com/horizoncd/horizon/pkg/registry/models"
)

type Registry struct {
	ID                    uint      `json:"id"`
	Name                  string    `json:"name"`
	Server                string    `json:"server"`
	Token                 string    `json:"token"`
	InsecureSkipTLSVerify bool      `json:"insecureSkipTLSVerify"`
	Kind                  string    `json:"kind"`
	Path                  string    `json:"path"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type Registries []*Registry

func ofRegistryModels(entities []*models.Registry) Registries {
	registries := make([]*Registry, 0)
	for _, entity := range entities {
		registries = append(registries, ofRegistryModel(entity))
	}

	return registries
}

func ofRegistryModel(entity *models.Registry) *Registry {
	return &Registry{
		ID:                    entity.ID,
		Name:                  entity.Name,
		Server:                entity.Server,
		Token:                 entity.Token,
		InsecureSkipTLSVerify: entity.InsecureSkipTLSVerify,
		Kind:                  entity.Kind,
		Path:                  entity.Path,
		CreatedAt:             entity.CreatedAt,
		UpdatedAt:             entity.UpdatedAt,
	}
}

type CreateRegistryRequest struct {
	Name                  string `json:"name"`
	Server                string `json:"server"`
	Token                 string `json:"token"`
	InsecureSkipTLSVerify bool   `json:"insecureSkipTLSVerify"`
	Path                  string `json:"path"`
	Kind                  string `json:"kind"`
}

type UpdateRegistryRequest struct {
	Name                  string `json:"name"`
	Server                string `json:"server"`
	Token                 string `json:"token"`
	InsecureSkipTLSVerify *bool  `json:"insecureSkipTLSVerify"`
	Path                  string `json:"path"`
	Kind                  string `json:"kind"`
}
