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

package kubeclient

import (
	"sync"

	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"k8s.io/client-go/rest"
)

var (
	ErrBuildKubeClientFailed = perror.New("build kubeclient failed")
)

var (
	Fty = NewFactory()
)

type Factory interface {
	GetByK8SServer(server, certificate string) (*rest.Config, *kube.Client, error)
}

type factory struct {
	cache *sync.Map
}

func NewFactory() Factory {
	return &factory{
		cache: &sync.Map{},
	}
}

type k8sClientCache struct {
	config *rest.Config
	client *kube.Client
}

func (f *factory) GetByK8SServer(server, certificate string) (*rest.Config, *kube.Client, error) {
	ret, ok := f.cache.Load(server)
	if ok {
		clientCache := ret.(*k8sClientCache)
		return clientCache.config, clientCache.client, nil
	}

	config, client, err := kube.BuildClientFromContent(certificate)
	if err != nil {
		return nil, nil, perror.Wrap(ErrBuildKubeClientFailed, err.Error())
	}

	f.cache.Store(server, &k8sClientCache{
		config: config,
		client: client,
	})

	return config, client, nil
}
