package harbor

import (
	"context"

	"g.hz.netease.com/horizon/pkg/harbor/manager"
	"g.hz.netease.com/horizon/pkg/harbor/models"
)

var (
	// Ctl global instance of the environment controller
	Ctl = NewController()
)

type Controller interface {
	// Create a harbor
	Create(ctx context.Context, request *CreateHarborRequest) (uint, error)
	// ListAll list all harbors
	ListAll(ctx context.Context) (Harbors, error)
}

func NewController() Controller {
	return &controller{harborManager: manager.Mgr}
}

type controller struct {
	harborManager manager.Manager
}

func (c controller) Create(ctx context.Context, request *CreateHarborRequest) (uint, error) {
	harbor, err := c.harborManager.Create(ctx, &models.Harbor{
		Server:          request.Server,
		Token:           request.Token,
		PreheatPolicyID: request.PreheatPolicyID,
	})
	if err != nil {
		return 0, err
	}

	return harbor.ID, nil
}

func (c controller) ListAll(ctx context.Context) (Harbors, error) {
	harbors, err := c.harborManager.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofHarborModels(harbors), nil
}
