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

package regioninformers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"

	"github.com/horizoncd/horizon/pkg/util/kube"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/region/manager"
	"github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type RegionClient struct {
	restConfig       *rest.Config
	regionID         uint
	watched          map[schema.GroupVersionResource]informers.GenericInformer
	dynamicWatched   map[schema.GroupVersionResource]informers.GenericInformer
	factory          informers.SharedInformerFactory
	dynamicFactory   dynamicinformer.DynamicSharedInformerFactory
	clientset        kubernetes.Interface
	dynamicClientset dynamic.Interface
	discoveryClient  *discovery.DiscoveryClient
	handlers         map[int]struct{}
	mapper           meta.RESTMapper
	stopCh           chan struct{}
	lastUpdated      time.Time
}

type Resource struct {
	MakeHandler func(uint, <-chan struct{}) (cache.ResourceEventHandler, error)
	GVR         schema.GroupVersionResource
}

type Target struct {
	Resource
	Regions map[uint]struct{}
}

// RegionInformers is a collection of informer factories for each region
type RegionInformers struct {
	regionMgr manager.Manager

	defaultResync time.Duration

	handlers []Resource

	// mu protects factories, lastUpdated and stopChanMap
	// mu      LoggingMutext
	clients sync.Map
}

// NewRegionInformers is called when initializing
func NewRegionInformers(regionMgr manager.Manager, defaultResync time.Duration) *RegionInformers {
	f := RegionInformers{
		regionMgr:     regionMgr,
		handlers:      make([]Resource, 0, 16),
		defaultResync: defaultResync,
	}
	regions, err := f.listRegion(context.Background())
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	for _, region := range regions {
		wg.Add(1)
		go func(region *models.Region) {
			log.Debugf(context.Background(), "Creating informer for region %s", region.Name)
			defer wg.Done()
			if err := f.NewRegionInformer(region); err != nil {
				log.Errorf(context.Background(), "Failed to create informer for region %s(%d): %v", region.Name, region.ID, err)
			}
		}(region)
	}
	wg.Wait()
	return &f
}

func (f *RegionInformers) NewRegionInformer(region *models.Region) error {
	if region == nil {
		return nil
	}

	config, err := clientcmd.NewClientConfigFromBytes([]byte(region.Certificate))
	if err != nil {
		return err
	}
	restConfig, err := config.ClientConfig()
	if err != nil {
		return err
	}
	restConfig = metadata.ConfigFor(restConfig)
	restConfig.QPS = kube.K8sClientConfigQPS
	restConfig.Burst = kube.K8sClientConfigBurst

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	discoveryClient := discovery.NewDiscoveryClient(clientSet.RESTClient())

	factory := informers.NewSharedInformerFactory(clientSet, f.defaultResync)

	dynamicClientSet, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	dynamicFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClientSet, f.defaultResync)

	resources, err := restmapper.GetAPIGroupResources(clientSet.Discovery())
	if err != nil {
		return err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(resources)

	stopCh := make(chan struct{}, 1)
	client := RegionClient{
		restConfig:       restConfig,
		regionID:         region.ID,
		watched:          make(map[schema.GroupVersionResource]informers.GenericInformer),
		dynamicWatched:   make(map[schema.GroupVersionResource]informers.GenericInformer),
		factory:          factory,
		dynamicFactory:   dynamicFactory,
		clientset:        clientSet,
		discoveryClient:  discoveryClient,
		dynamicClientset: dynamicClientSet,
		handlers:         make(map[int]struct{}),
		mapper:           mapper,
		stopCh:           stopCh,
		lastUpdated:      region.UpdatedAt,
	}

	f.registerHandler(&client)

	f.clients.Store(region.ID, &client)
	return nil
}

func (f *RegionInformers) DeleteRegionInformer(regionID uint) {
	client, existed := f.clients.LoadAndDelete(regionID)
	if existed {
		regionClient, ok := client.(*RegionClient)
		if ok && regionClient != nil && regionClient.stopCh != nil {
			close(regionClient.stopCh)
		}
	}
}

// listRegion list all regions which are not disabled
func (f *RegionInformers) listRegion(ctx context.Context) ([]*models.Region, error) {
	regions, err := f.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

// WatchRegion blocks until ctx is done, and watch the database for region changes
func (f *RegionInformers) WatchRegion(ctx context.Context, pollInterval time.Duration) {
	err := wait.Poll(pollInterval, 0, func() (done bool, err error) {
		select {
		case <-ctx.Done():
			return true, nil
		default:
		}

		f.watchDB()
		return false, nil
	})
	if err != nil {
		log.Errorf(ctx, "WatchRegion polling error: %v", err)
	}
}

// watchDB watches the database and fetches regions diff with the cache
func (f *RegionInformers) watchDB() {
	created, updated, deleted, err := f.diffRegion()
	if err != nil {
		log.Errorf(context.Background(), "diffRegion error: %v", err)
		return
	}

	created = append(created, updated...)
	deleted = append(deleted, updated...)

	for _, region := range deleted {
		f.DeleteRegionInformer(region.ID)
	}

	for _, region := range created {
		if err := f.NewRegionInformer(region); err != nil {
			log.Errorf(context.Background(), "NewRegionInformer error: %v", err)
		}
	}
}

func (f *RegionInformers) diffRegion() ([]*models.Region, []*models.Region, []*models.Region, error) {
	regions, err := f.listRegion(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}

	created, updated, deleted := make([]*models.Region, 0), make([]*models.Region, 0), make([]*models.Region, 0)

	regionsID := make(map[uint]*models.Region)
	for _, region := range regions {
		regionsID[region.ID] = region
		client, existed := f.clients.Load(region.ID)
		if !existed {
			created = append(created, region)
			continue
		}
		if regionClient, ok := client.(*RegionClient); ok && regionClient != nil {
			if region.UpdatedAt.After(regionClient.lastUpdated) {
				updated = append(updated, region)
			}
		}
	}

	f.clients.Range(func(key, value interface{}) bool {
		regionID := key.(uint)
		if _, ok := regionsID[regionID]; !ok {
			deleted = append(deleted, regionsID[regionID])
		}
		return true
	})

	return created, updated, deleted, nil
}

type ClientSetOperation func(clientset kubernetes.Interface) error

// GetClientSet gets the client for the given region
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetClientSet(regionID uint, operation ClientSetOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return operation(client.(*RegionClient).clientset)
}

func (f *RegionInformers) GetRestConfig(regionID uint) (*rest.Config, error) {
	if err := f.ensureRegion(regionID); err != nil {
		return nil, err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return nil, herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return client.(*RegionClient).restConfig, nil
}

type DynamicClientSetOperation func(clientset dynamic.Interface) error

// GetDynamicClientSet gets the dynamic clientset for the given region
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetDynamicClientSet(regionID uint, operation DynamicClientSetOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return operation(client.(*RegionClient).dynamicClientset)
}

func (f *RegionInformers) whetherGVRExist(regionID uint, gvr schema.GroupVersionResource) bool {
	client, ok := f.clients.Load(regionID)
	if !ok {
		return false
	}
	_, ok = client.(*RegionClient).watched[gvr]
	return ok
}

func (f *RegionInformers) watchGVR(regionID uint, gvr schema.GroupVersionResource) error {
	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	informer, err := regionClient.factory.ForResource(gvr)
	if err != nil {
		return err
	}
	regionClient.watched[gvr] = informer

	regionClient.factory.Start(regionClient.stopCh)
	regionClient.factory.WaitForCacheSync(regionClient.stopCh)

	return nil
}

func (f *RegionInformers) watchDynamicGvr(regionID uint, gvr schema.GroupVersionResource) error {
	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient.dynamicWatched[gvr] = regionClient.dynamicFactory.ForResource(gvr)
	regionClient.dynamicFactory.Start(regionClient.stopCh)
	regionClient.dynamicFactory.WaitForCacheSync(regionClient.stopCh)
	return nil
}

func (f *RegionInformers) ensureGVR(regionID uint, gvr schema.GroupVersionResource) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	if !f.whetherGVRExist(regionID, gvr) {
		return f.watchGVR(regionID, gvr)
	}
	return nil
}

func (f *RegionInformers) ensureDynamicGVR(regionID uint, gvr schema.GroupVersionResource) error {
	if !f.whetherGVRExist(regionID, gvr) {
		client, ok := f.clients.Load(regionID)
		if !ok {
			return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
		}
		regionClient, ok := client.(*RegionClient)
		if !ok {
			return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
		}
		if !f.resourceExistInK8S(gvr, regionClient) {
			return fmt.Errorf("resource %s not exist in region %d", gvr.String(), regionID)
		}
		return f.watchDynamicGvr(regionID, gvr)
	}
	return nil
}

type InformerOperation func(informer informers.GenericInformer) error

// GetInformer gets the informer for the given region and gvr
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetInformer(regionID uint,
	gvr schema.GroupVersionResource, operation InformerOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	if err := f.ensureGVR(regionID, gvr); err != nil {
		return err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	informer, err := regionClient.factory.ForResource(gvr)
	if err != nil {
		return err
	}
	return operation(informer)
}

type FactoryOperation func(factory informers.SharedInformerFactory) error

func (f *RegionInformers) GetFactory(regionID uint, operation FactoryOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return operation(regionClient.factory)
}

type DynamicInformerOperation func(informer informers.GenericInformer) error

// GetDynamicInformer gets the informer for the given region and gvr
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetDynamicInformer(regionID uint,
	gvr schema.GroupVersionResource, operation DynamicInformerOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	if err := f.ensureDynamicGVR(regionID, gvr); err != nil {
		return err
	}

	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return operation(regionClient.dynamicFactory.ForResource(gvr))
}

type DynamicFactoryOperation func(factory dynamicinformer.DynamicSharedInformerFactory) error

func (f *RegionInformers) GetDynamicFactory(regionID uint, operation DynamicFactoryOperation) error {
	// 确定这个 region 的 factory 是否存在，如果不存在则创建
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	// TODO: check this
	log.Debugf(context.Background(), "got dynamic factory: %d", regionID)
	client, ok := f.clients.Load(regionID)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	return operation(regionClient.dynamicFactory)
}

func (f *RegionInformers) checkExist(regionID uint) bool {
	_, ok := f.clients.Load(regionID)
	return ok
}

// ensureRegion runs the operation if the region exists, otherwise create a new factory
func (f *RegionInformers) ensureRegion(regionID uint) error {
	if !f.checkExist(regionID) {
		region, err := f.regionMgr.GetRegionByID(context.Background(), regionID)
		if err != nil {
			return err
		}
		return f.NewRegionInformer(region.Region)
	}
	return nil
}

func (f *RegionInformers) Register(handlers ...Resource) {
	f.handlers = append(f.handlers, handlers...)

	f.clients.Range(func(key, value interface{}) bool {
		client := value.(*RegionClient)
		f.registerHandler(client)
		return true
	})
}

func (f *RegionInformers) resourceExistInK8S(gvr schema.GroupVersionResource, client *RegionClient) bool {
	if client == nil {
		return false
	}

	apiResourceList, err := client.discoveryClient.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
	if err != nil {
		log.Errorf(context.Background(), "list api resources failed: %v", err)
		return false
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Name == gvr.Resource {
			return true
		}
	}

	return false
}

func (f *RegionInformers) registerHandler(client *RegionClient) {
	for i, handler := range f.handlers {
		if _, ok := client.handlers[i]; ok {
			continue
		}

		if !f.resourceExistInK8S(handler.GVR, client) {
			continue
		}

		informer := client.dynamicFactory.ForResource(handler.GVR)
		if handler.MakeHandler != nil {
			eventHandler, err := handler.MakeHandler(client.regionID, client.stopCh)
			if err != nil {
				log.Errorf(context.Background(), "make handler for %s failed: %v", handler.GVR, err)
				continue
			}
			informer.Informer().AddEventHandler(eventHandler)
		}
		client.handlers[i] = struct{}{}
	}

	client.factory.Start(client.stopCh)
	client.dynamicFactory.Start(client.stopCh)
}

func (f *RegionInformers) Start(regionIDs ...uint) error {
	if len(regionIDs) == 0 {
		regions, err := f.listRegion(context.Background())
		if err != nil {
			return err
		}
		for _, region := range regions {
			regionIDs = append(regionIDs, region.ID)
		}
	}

	for _, regionID := range regionIDs {
		if err := f.ensureRegion(regionID); err != nil {
			return err
		}
	}

	f.clients.Range(func(key, value interface{}) bool {
		client := value.(*RegionClient)
		stopCh := client.stopCh
		client.factory.Start(stopCh)
		client.dynamicFactory.Start(stopCh)
		return true
	})
	return nil
}

func (f *RegionInformers) GVK2GVR(regionID uint, GVK schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	client, ok := f.clients.Load(regionID)
	if !ok {
		return schema.GroupVersionResource{},
			herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}
	regionClient, ok := client.(*RegionClient)
	if !ok {
		return schema.GroupVersionResource{},
			herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}

	mapping, err := regionClient.mapper.RESTMapping(GVK.GroupKind(), GVK.Version)
	if err != nil {
		return schema.GroupVersionResource{},
			herrors.NewErrNotFound(herrors.ResourceInK8S, fmt.Sprintf("mapping for %s: %s", GVK, err))
	}
	return mapping.Resource, nil
}
