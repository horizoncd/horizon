package regioninformers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"

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
	mu      sync.RWMutex
	clients map[uint]*RegionClient
}

// NewRegionInformers is called when initializing
func NewRegionInformers(regionMgr manager.Manager, defaultResync time.Duration) *RegionInformers {
	f := RegionInformers{
		regionMgr:     regionMgr,
		clients:       make(map[uint]*RegionClient),
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
			if err := f.NewRegionInformers(region); err != nil {
				log.Errorf(context.Background(), "Failed to create informer for region %s(%d): %v", region.Name, region.ID, err)
			}
		}(region)
	}
	wg.Wait()
	return &f
}

func (f *RegionInformers) Close() {
	for _, client := range f.clients {
		close(client.stopCh)
	}
}

func (f *RegionInformers) NewRegionInformers(region *models.Region) error {
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

	f.mu.Lock()
	defer f.mu.Unlock()
	f.clients[region.ID] = &client

	f.registerHandler(&client)
	return nil
}

func (f *RegionInformers) DeleteRegionInformer(regionID uint) {
	f.mu.Lock()
	defer f.mu.Unlock()
	client, ok := f.clients[regionID]
	if !ok {
		return
	}
	close(client.stopCh)
	delete(f.clients, regionID)
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
		if err := f.NewRegionInformers(region); err != nil {
			log.Errorf(context.Background(), "NewRegionInformers error: %v", err)
		}
	}
}

func (f *RegionInformers) diffRegion() ([]*models.Region, []*models.Region, []*models.Region, error) {
	regions, err := f.listRegion(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}

	created, updated, deleted := make([]*models.Region, 0), make([]*models.Region, 0), make([]*models.Region, 0)

	f.mu.RLock()
	defer f.mu.RUnlock()

	regionsID := make(map[uint]*models.Region)
	for _, region := range regions {
		regionsID[region.ID] = region
		if _, ok := f.clients[region.ID]; !ok {
			created = append(created, region)
			continue
		}
		if region.UpdatedAt.After(f.clients[region.ID].lastUpdated) {
			updated = append(updated, region)
		}
	}

	for regionID := range f.clients {
		if region, ok := regionsID[regionID]; !ok {
			deleted = append(deleted, region)
		}
	}

	return created, updated, deleted, nil
}

type ClientSetOperation func(clientset kubernetes.Interface) error

// GetClientSet gets the client for the given region
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetClientSet(regionID uint, operation ClientSetOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return operation(f.clients[regionID].clientset)
}

func (f *RegionInformers) GetRestConfig(regionID uint) (*rest.Config, error) {
	if err := f.ensureRegion(regionID); err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.clients[regionID].restConfig, nil
}

type DynamicClientSetOperation func(clientset dynamic.Interface) error

// GetDynamicClientSet gets the dynamic clientset for the given region
// it should have no reference for factory outside RegionInformers
func (f *RegionInformers) GetDynamicClientSet(regionID uint, operation DynamicClientSetOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return operation(f.clients[regionID].dynamicClientset)
}

func (f *RegionInformers) whetherGVRExist(regionID uint, gvr schema.GroupVersionResource) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.clients[regionID].watched[gvr]
	return ok
}

func (f *RegionInformers) watchGVR(regionID uint, gvr schema.GroupVersionResource) error {
	f.mu.Lock()
	client := f.clients[regionID]
	informer, err := client.factory.ForResource(gvr)
	if err != nil {
		return err
	}
	client.watched[gvr] = informer
	f.mu.Unlock()

	client.factory.Start(client.stopCh)
	client.factory.WaitForCacheSync(client.stopCh)
	return nil
}

func (f *RegionInformers) watchDynamicGvr(regionID uint, gvr schema.GroupVersionResource) {
	f.mu.Lock()
	client := f.clients[regionID]
	informer := client.dynamicFactory.ForResource(gvr)
	client.dynamicWatched[gvr] = informer
	f.mu.Unlock()

	client.dynamicFactory.Start(client.stopCh)
	client.dynamicFactory.WaitForCacheSync(client.stopCh)
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
		if !f.resourceExistInK8S(gvr, f.clients[regionID]) {
			return fmt.Errorf("resource %s not exist in region %d", gvr.String(), regionID)
		}
		f.watchDynamicGvr(regionID, gvr)
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

	f.mu.RLock()
	defer f.mu.RUnlock()
	informer, err := f.clients[regionID].factory.ForResource(gvr)
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

	f.mu.RLock()
	defer f.mu.RUnlock()
	return operation(f.clients[regionID].factory)
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

	f.mu.RLock()
	defer f.mu.RUnlock()
	return operation(f.clients[regionID].dynamicFactory.ForResource(gvr))
}

type DynamicFactoryOperation func(factory dynamicinformer.DynamicSharedInformerFactory) error

func (f *RegionInformers) GetDynamicFactory(regionID uint, operation DynamicFactoryOperation) error {
	if err := f.ensureRegion(regionID); err != nil {
		return err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return operation(f.clients[regionID].dynamicFactory)
}

func (f *RegionInformers) checkExist(regionID uint) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.clients[regionID]
	return ok
}

// ensureRegion runs the operation if the region exists, otherwise create a new factory
func (f *RegionInformers) ensureRegion(regionID uint) error {
	if !f.checkExist(regionID) {
		region, err := f.regionMgr.GetRegionByID(context.Background(), regionID)
		if err != nil {
			return err
		}
		return f.NewRegionInformers(region.Region)
	}
	return nil
}

func (f *RegionInformers) Register(handlers ...Resource) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers = append(f.handlers, handlers...)

	for _, client := range f.clients {
		f.registerHandler(client)
	}
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

	for _, regionID := range regionIDs {
		func() {
			f.mu.RLock()
			defer f.mu.RUnlock()

			client := f.clients[regionID]
			stopCh := client.stopCh
			client.dynamicFactory.Start(stopCh)
			client.factory.Start(stopCh)
		}()
	}
	return nil
}

func (f *RegionInformers) GVK2GVR(regionID uint, GVK schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	client, ok := f.clients[regionID]
	if !ok {
		return schema.GroupVersionResource{},
			herrors.NewErrNotFound(herrors.RegionInDB, fmt.Sprintf("region %d", regionID))
	}

	mapping, err := client.mapper.RESTMapping(GVK.GroupKind(), GVK.Version)
	if err != nil {
		return schema.GroupVersionResource{},
			herrors.NewErrNotFound(herrors.ResourceInK8S, fmt.Sprintf("mapping for %s: %s", GVK, err))
	}
	return mapping.Resource, nil
}
