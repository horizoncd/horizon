package kubeclient

import (
	"context"
	"sync"

	perror "g.hz.netease.com/horizon/pkg/errors"
	k8sclustermanager "g.hz.netease.com/horizon/pkg/k8scluster/manager"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"k8s.io/client-go/rest"
)

var (
	ErrGetClusterFailed      = perror.New("get cluster failed")
	ErrBuildKubeClientFailed = perror.New("build kubeclient failed")
)

var (
	Fty = NewFactory()
)

type Factory interface {
	GetByK8SServer(ctx context.Context, server string) (*rest.Config, *kube.Client, error)
}

type factory struct {
	k8sClusterMgr k8sclustermanager.Manager
	cache         *sync.Map
}

func NewFactory() Factory {
	return &factory{
		k8sClusterMgr: k8sclustermanager.Mgr,
		cache:         &sync.Map{},
	}
}

type k8sClientCache struct {
	config *rest.Config
	client *kube.Client
}

func (f *factory) GetByK8SServer(ctx context.Context, server string) (*rest.Config, *kube.Client, error) {
	ret, ok := f.cache.Load(server)
	if ok {
		clientCache := ret.(*k8sClientCache)
		return clientCache.config, clientCache.client, nil
	}

	k8sCluster, err := f.k8sClusterMgr.GetByServer(ctx, server)
	if err != nil {
		return nil, nil, perror.Wrap(ErrGetClusterFailed, err.Error())
	}

	config, client, err := kube.BuildClientFromContent(k8sCluster.Certificate)
	if err != nil {
		return nil, nil, perror.Wrap(ErrBuildKubeClientFailed, err.Error())
	}

	f.cache.Store(server, &k8sClientCache{
		config: config,
		client: client,
	})

	return config, client, nil
}
