package argocd

import (
	"fmt"
	"sync"

	"g.hz.netease.com/horizon/pkg/config/argocd"
)

type Factory interface {
	GetArgoCD(environment string) (ArgoCD, error)
}

type factory struct {
	cache *sync.Map
}

func NewFactory(argoCDMapper argocd.Mapper) Factory {
	cache := &sync.Map{}
	for env, argoCDConf := range argoCDMapper {
		argoCD := NewArgoCD(argoCDConf.URL, argoCDConf.Token)
		cache.Store(env, argoCD)
	}
	return &factory{
		cache: cache,
	}
}

func (f *factory) GetArgoCD(environment string) (ArgoCD, error) {
	var ret interface{}
	var ok bool
	if ret, ok = f.cache.Load(environment); !ok {
		return nil, fmt.Errorf("the argoCD for environment: %v is not found", environment)
	}
	return ret.(ArgoCD), nil
}
