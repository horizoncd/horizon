package factory

import (
	"context"
	"fmt"
	"sync"

	"g.hz.netease.com/horizon/pkg/cluster/registry"
)

var (
	// Fty is the global registry factory
	Fty = NewFactory()
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../../mock/pkg/cluster/registry/factory/factory_mock.go -package=mock_factory
type Factory interface {
	GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error)
}

type factory struct {
	registryCache *sync.Map
}

func NewFactory() Factory {
	registryCache := &sync.Map{}
	return &factory{
		registryCache: registryCache,
	}
}

func (f *factory) GetRegistryByConfig(ctx context.Context, config *registry.Config) (registry.Registry, error) {
	key := fmt.Sprintf("%v-%v-%v-%v", config.Server, config.Token, config.Path, config.Kind)
	if ret, ok := f.registryCache.Load(key); ok {
		return ret.(registry.Registry), nil
	}
	rg, err := registry.NewRegistry(config)
	if err != nil {
		return nil, err
	}
	f.registryCache.Store(key, rg)
	return rg, nil
}
