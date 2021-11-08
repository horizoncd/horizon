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

type Factory interface {
	GetByHarborConfig(ctx context.Context, harbor *registry.HarborConfig) registry.Registry
}

type factory struct {
	harborRegistryCache *sync.Map
}

func NewFactory() Factory {
	harborRegistryCache := &sync.Map{}
	return &factory{
		harborRegistryCache: harborRegistryCache,
	}
}

func (f *factory) GetByHarborConfig(ctx context.Context, harbor *registry.HarborConfig) registry.Registry {
	key := fmt.Sprintf("%v-%v", harbor.Server, harbor.Token)
	if ret, ok := f.harborRegistryCache.Load(key); ok {
		return ret.(registry.Registry)
	}
	harborRegistry := registry.NewHarborRegistry(harbor)
	f.harborRegistryCache.Store(key, harborRegistry)
	return harborRegistry
}
