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
	"context"
	"sync"

	"github.com/horizoncd/horizon/pkg/cluster/registry"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/registry/manager"
	"github.com/horizoncd/horizon/pkg/registry/models"
)

var kindCache []string

type Controller interface {
	// Create a registry
	Create(ctx context.Context, request *CreateRegistryRequest) (uint, error)
	// ListAll list all registries
	ListAll(ctx context.Context) (Registries, error)
	// UpdateByID update a registry
	UpdateByID(ctx context.Context, id uint, request *UpdateRegistryRequest) error
	// DeleteByID delete a registry by id
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Registry, error)
	GetKinds(ctx context.Context) []string
}

func NewController(param *param.Param) Controller {
	return &controller{registryManager: param.RegistryManager}
}

type controller struct {
	registryManager manager.Manager
}

func (c controller) Create(ctx context.Context, request *CreateRegistryRequest) (uint, error) {
	id, err := c.registryManager.Create(ctx, &models.Registry{
		Name:                  request.Name,
		Server:                request.Server,
		Token:                 request.Token,
		InsecureSkipTLSVerify: request.InsecureSkipTLSVerify,
		Path:                  request.Path,
		Kind:                  request.Kind,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c controller) ListAll(ctx context.Context) (Registries, error) {
	registries, err := c.registryManager.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegistryModels(registries), nil
}

func (c controller) UpdateByID(ctx context.Context, id uint, request *UpdateRegistryRequest) error {
	_, err := c.registryManager.GetByID(ctx, id)
	if err != nil {
		return err
	}

	registry := &models.Registry{
		Name:   request.Name,
		Server: request.Server,
		Token:  request.Token,
		Kind:   request.Kind,
		Path:   request.Path,
	}

	if request.InsecureSkipTLSVerify != nil {
		registry.InsecureSkipTLSVerify = *request.InsecureSkipTLSVerify
	}

	err = c.registryManager.UpdateByID(ctx, id, registry)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) DeleteByID(ctx context.Context, id uint) error {
	err := c.registryManager.DeleteByID(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) GetByID(ctx context.Context, id uint) (*Registry, error) {
	registry, err := c.registryManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofRegistryModel(registry), nil
}

func (c controller) GetKinds(ctx context.Context) []string {
	var once sync.Once
	once.Do(func() {
		kindCache = registry.GetKinds()
	})

	return kindCache
}
