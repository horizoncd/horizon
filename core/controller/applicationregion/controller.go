package applicationregion

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/sets"
)

type Controller interface {
	List(ctx context.Context, applicationID uint) (ApplicationRegion, error)
	Update(ctx context.Context, applicationID uint, applicationRegionMap map[string]string) error
}

type controller struct {
	regionConfig *region.Config

	mgr            manager.Manager
	regionMgr      regionmanager.Manager
	environmentMgr envmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(regionConfig *region.Config) Controller {
	return &controller{
		regionConfig:   regionConfig,
		mgr:            manager.Mgr,
		regionMgr:      regionmanager.Mgr,
		environmentMgr: envmanager.Mgr,
	}
}

func (c *controller) List(ctx context.Context, applicationID uint) (ApplicationRegion, error) {
	applicationRegions, err := c.mgr.ListByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, perrors.WithMessage(err, "failed to list application regions")
	}

	environments, err := c.environmentMgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, perrors.WithMessage(err, "failed to list environment")
	}

	regions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, perrors.WithMessage(err, "failed to list region")
	}

	return ofApplicationRegion(applicationRegions, regions, environments, c.regionConfig), nil
}

func (c *controller) Update(ctx context.Context, applicationID uint, applicationRegionMap map[string]string) error {
	applicationRegions := make([]*models.ApplicationRegion, 0)
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return perrors.WithMessage(err, "no user in context")
	}

	// get all environment
	allEnvironment, err := c.environmentMgr.ListAllEnvironment(ctx)
	if err != nil {
		return perrors.WithMessage(err, "failed to list all environment")
	}
	environmentSet := sets.NewString()
	for _, env := range allEnvironment {
		environmentSet.Insert(env.Name)
	}

	// get all region
	allRegions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return perrors.WithMessage(err, "failed to list all regions")
	}
	regionSet := sets.NewString()
	for _, region := range allRegions {
		regionSet.Insert(region.Name)
	}

	for k, v := range applicationRegionMap {
		if !environmentSet.Has(k) {
			return fmt.Errorf("environment %s is not exists", k)
		}
		if !regionSet.Has(v) {
			return fmt.Errorf("region %s is not exists", v)
		}
		applicationRegions = append(applicationRegions, &models.ApplicationRegion{
			ApplicationID:   applicationID,
			EnvironmentName: k,
			RegionName:      v,
			CreatedBy:       currentUser.GetID(),
			UpdatedBy:       currentUser.GetID(),
		})
	}

	return c.mgr.UpsertByApplicationID(ctx, applicationID, applicationRegions)
}
