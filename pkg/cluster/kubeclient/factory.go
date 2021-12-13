package kubeclient

import (
	"context"
	"fmt"
	"sync"

	k8sclustermanager "g.hz.netease.com/horizon/pkg/k8scluster/manager"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	Fty = NewFactory()
)

type Factory interface {
	GetByK8SServer(ctx context.Context, server string) (*rest.Config, kubernetes.Interface, error)
	GetDynamicByK8SServer(ctx context.Context, server string) (*rest.Config, dynamic.Interface, error)
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
	client kubernetes.Interface
}

type k8sDynamicClientCache struct {
	config *rest.Config
	client dynamic.Interface
}

func (f *factory) GetByK8SServer(ctx context.Context, server string) (*rest.Config, kubernetes.Interface, error) {
	ret, ok := f.cache.Load(server)
	if ok {
		clientCache := ret.(*k8sClientCache)
		return clientCache.config, clientCache.client, nil
	}

	k8sCluster, err := f.k8sClusterMgr.GetByServer(ctx, server)
	if err != nil {
		return nil, nil, err
	}

	config, client, err := kube.BuildClientFromContent(k8sCluster.Certificate)
	if err != nil {
		return nil, nil, err
	}

	f.cache.Store(server, &k8sClientCache{
		config: config,
		client: client,
	})

	return config, client, nil
}

func getK8sDynamicClientName(server string) string {
	return fmt.Sprintf(`%s-dynamic`, server)
}
func (f *factory) GetDynamicByK8SServer(ctx context.Context, server string) (*rest.Config, dynamic.Interface, error) {
	clientName := getK8sDynamicClientName(server)
	ret, ok := f.cache.Load(clientName)
	if ok {
		clientCache := ret.(*k8sDynamicClientCache)
		return clientCache.config, clientCache.client, nil
	}

	k8sCluster, err := f.k8sClusterMgr.GetByServer(ctx, server)
	if err != nil {
		return nil, nil, err
	}

	config, client, err := kube.BuildDynamicClientFromContent(k8sCluster.Certificate)
	if err != nil {
		return nil, nil, err
	}

	f.cache.Store(clientName, &k8sDynamicClientCache{
		config: config,
		client: client,
	})

	return config, client, nil
}
