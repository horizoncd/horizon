package harbor

import (
	"context"

	"g.hz.netease.com/horizon/pkg/harbor/manager"
	"g.hz.netease.com/horizon/pkg/harbor/models"
	"g.hz.netease.com/horizon/pkg/param"
)

type Controller interface {
	// Create a harbor
	Create(ctx context.Context, request *CreateHarborRequest) (uint, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) (Harbors, error)
	// UpdateByID update a harbor
	UpdateByID(ctx context.Context, id uint, request *UpdateHarborRequest) error
	// DeleteByID delete a harbor by id
	DeleteByID(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Harbor, error)
}

func NewController(param *param.Param) Controller {
	return &controller{harborManager: param.HarborManager}
}

type controller struct {
	harborManager manager.Manager
}

func (c controller) Create(ctx context.Context, request *CreateHarborRequest) (uint, error) {
	id, err := c.harborManager.Create(ctx, &models.Harbor{
		Name:            request.Name,
		Server:          request.Server,
		Token:           request.Token,
		PreheatPolicyID: request.PreheatPolicyID,
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c controller) ListAll(ctx context.Context) (Harbors, error) {
	harbors, err := c.harborManager.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofHarborModels(harbors), nil
}

func (c controller) UpdateByID(ctx context.Context, id uint, request *UpdateHarborRequest) error {
	err := c.harborManager.UpdateByID(ctx, id, &models.Harbor{
		Name:            request.Name,
		Server:          request.Server,
		Token:           request.Token,
		PreheatPolicyID: request.PreheatPolicyID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c controller) DeleteByID(ctx context.Context, id uint) error {
	err := c.harborManager.DeleteByID(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) GetByID(ctx context.Context, id uint) (*Harbor, error) {
	harbor, err := c.harborManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ofHarborModel(harbor), nil
}
