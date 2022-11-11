package registry

import (
	"context"
	"sync"

	"g.hz.netease.com/horizon/pkg/cluster/registry"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/registry/manager"
	"g.hz.netease.com/horizon/pkg/registry/models"
)

var kindCache []string

type Controller interface {
	// Create a registry
	Create(ctx context.Context, request *CreateRegistryRequest) (uint, error)
	// ListAll list all registries
	ListAll(ctx context.Context) (Registries, error)
	// UpdateByID update a registry
	UpdateByID(ctx context.Context, id uint, request *UpdateRegistryRequest) error
	// DeleteByID delete a registry by id
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Registry, error)
	GetKinds(ctx context.Context) []string
}

func NewController(param *param.Param) Controller {
	return &controller{registryManager: param.RegistryManager}
}

type controller struct {
	registryManager manager.Manager
}

func (c controller) Create(ctx context.Context, request *CreateRegistryRequest) (uint, error) {
	id, err := c.registryManager.Create(ctx, &models.Registry{
		Name:                  request.Name,
		Server:                request.Server,
		Token:                 request.Token,
		InsecureSkipTLSVerify: request.InsecureSkipTLSVerify,
		Path:                  request.Path,
		Kind:                  request.Kind,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c controller) ListAll(ctx context.Context) (Registries, error) {
	registries, err := c.registryManager.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegistryModels(registries), nil
}

func (c controller) UpdateByID(ctx context.Context, id uint, request *UpdateRegistryRequest) error {
	_, err := c.registryManager.GetByID(ctx, id)
	if err != nil {
		return err
	}

	registry := &models.Registry{
		Name:   request.Name,
		Server: request.Server,
		Token:  request.Token,
		Kind:   request.Kind,
		Path:   request.Path,
	}

	if request.InsecureSkipTLSVerify != nil {
		registry.InsecureSkipTLSVerify = *request.InsecureSkipTLSVerify
	}

	err = c.registryManager.UpdateByID(ctx, id, registry)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) DeleteByID(ctx context.Context, id uint) error {
	err := c.registryManager.DeleteByID(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) GetByID(ctx context.Context, id uint) (*Registry, error) {
	registry, err := c.registryManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofRegistryModel(registry), nil
}

func (c controller) GetKinds(ctx context.Context) []string {
	var once sync.Once
	once.Do(func() {
		kindCache = registry.GetKinds()
	})

	return kindCache
}
