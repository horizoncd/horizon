package factory

import (
	"context"
	"fmt"
	"sync"

	"github.com/horizoncd/horizon/pkg/cluster/registry"
)

var (
	// Fty is the global registry factory
	Fty = NewRegistryCache()
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../../mock/pkg/cluster/registry/factory/factory_mock.go -package=mock_factory
type RegistryGetter interface {
	GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error)
}

type RegistryCache struct {
	*sync.Map
}

func NewRegistryCache() *RegistryCache {
	return &RegistryCache{
		&sync.Map{},
	}
}

func (f *RegistryCache) GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error) {
	key := fmt.Sprintf("%v-%v-%v-%v", config.Server, config.Token, config.Path, config.Kind)
	if ret, ok := f.Load(key); ok {
		return ret.(registry.Registry), nil
	}
	rg, err := registry.NewRegistry(config)
	if err != nil {
		return nil, err
	}
	f.Store(key, rg)
	return rg, nil
}
