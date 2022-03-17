package argocd

import (
	"fmt"
	"sync"

	herrors "g.hz.netease.com/horizon/core/errors"
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
		return nil, herrors.NewErrNotFound(herrors.ArgoCD, fmt.Sprintf("argo cd not found for environment %s", environment))
	}
	return ret.(ArgoCD), nil
}
