package argocd

import (
	"sync"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/argocd"
)

const _default = "default"

type Factory interface {
	GetArgoCD(environment string) (ArgoCD, error)
}

type factory struct {
	cache *sync.Map
}

func NewFactory(argoCDMapper argocd.Mapper) Factory {
	cache := &sync.Map{}
	for env, argoCDConf := range argoCDMapper {
		argoCD := NewArgoCD(argoCDConf.URL, argoCDConf.Token, argoCDConf.Namespace)
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
		// check and use default cd
		if ret, ok = f.cache.Load(_default); !ok {
			return nil, herrors.NewErrNotFound(herrors.ArgoCD, "default argo cd not found")
		}
	}
	return ret.(ArgoCD), nil
}
