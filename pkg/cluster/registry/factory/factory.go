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

package factory

import (
	"context"
	"fmt"
	"sync"

	"github.com/horizoncd/horizon/pkg/cluster/registry"
)

var (
	// Fty is the global registry factory
	Fty = newRegistryCache()
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../../mock/pkg/cluster/registry/factory/factory_mock.go -package=mock_factory
type RegistryGetter interface {
	GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error)
}

type registryCache struct {
	*sync.Map
}

func newRegistryCache() *registryCache {
	return &registryCache{
		&sync.Map{},
	}
}

func (r *registryCache) GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error) {
	key := fmt.Sprintf("%v-%v-%v-%v", config.Server, config.Token, config.Path, config.Kind)
	if ret, ok := r.Load(key); ok {
		return ret.(registry.Registry), nil
	}
	rg, err := registry.NewRegistry(config)
	if err != nil {
		return nil, err
	}
	r.Store(key, rg)
	return rg, nil
}
