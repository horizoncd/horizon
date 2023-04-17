package kubeclient

import (
	"sync"

	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"k8s.io/client-go/rest"
)

var ErrBuildKubeClientFailed = perror.New("build kubeclient failed")

var Fty = NewFactory()

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
