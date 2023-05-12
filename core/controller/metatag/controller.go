package metatag

import (
	"context"
	metatagmanager "github.com/horizoncd/horizon/pkg/metatag/manager"
	"github.com/horizoncd/horizon/pkg/metatag/models"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	CreateMetatags(ctx context.Context, createMetatagRequest *CreateMetatagsRequest) error
	GetMetatagKeys(ctx context.Context) ([]string, error)
	GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error)
}

type controller struct {
	manager metatagmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		manager: param.MetatagManager,
	}
}

func (c controller) CreateMetatags(ctx context.Context, createMetatagRequest *CreateMetatagsRequest) error {
	return c.manager.CreateMetatags(ctx, createMetatagRequest.Metatags)
}

func (c controller) GetMetatagKeys(ctx context.Context) ([]string, error) {
	return c.manager.GetMetatagKeys(ctx)
}

func (c controller) GetMetatagsByKey(ctx context.Context, key string) ([]*models.Metatag, error) {
	return c.manager.GetMetatagsByKey(ctx, key)
}
