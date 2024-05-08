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

package argocd

import (
	"sync"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/config/argocd"
)

const _default = "default"

type Factory interface {
	GetArgoCD(region string, environment string) (ArgoCD, error)
}

type factory struct {
	cache       *sync.Map
	regionCache *sync.Map
}

func NewFactory(argoCDMapper argocd.Mapper, regionArgoCDMapper argocd.RegionMapper) Factory {
	cache := &sync.Map{}
	for env, argoCDConf := range argoCDMapper {
		argoCD := NewArgoCD(argoCDConf.URL, argoCDConf.Token, argoCDConf.Namespace)
		cache.Store(env, argoCD)
	}
	regionCache := &sync.Map{}
	for region, argoCDConf := range regionArgoCDMapper {
		argoCD := NewArgoCD(argoCDConf.URL, argoCDConf.Token, argoCDConf.Namespace)
		regionCache.Store(region, argoCD)
	}
	return &factory{
		cache:       cache,
		regionCache: regionCache,
	}
}

func (f *factory) GetArgoCD(region string, environment string) (ArgoCD, error) {
	var ret interface{}
	var ok bool
	if ret, ok = f.regionCache.Load(region); !ok {
		if ret, ok = f.cache.Load(environment); !ok {
			// check and use default cd
			if ret, ok = f.cache.Load(_default); !ok {
				return nil, herrors.NewErrNotFound(herrors.ArgoCD, "default argo cd not found")
			}
		}
	}
	return ret.(ArgoCD), nil
}
