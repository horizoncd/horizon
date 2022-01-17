package applicationregion

import (
	"context"
	"errors"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/sets"
)

var (
	ErrEnvironmentNotFound = errors.New("environment not found")
	ErrRegionNotFound      = errors.New("region not found")
)

type Controller interface {
	List(ctx context.Context, applicationID uint) (ApplicationRegion, error)
	Update(ctx context.Context, applicationID uint, regions ApplicationRegion) error
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

	return ofApplicationRegion(applicationRegions, environments, c.regionConfig), nil
}

func (c *controller) Update(ctx context.Context, applicationID uint, regions ApplicationRegion) error {
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
	for _, r := range allRegions {
		regionSet.Insert(r.Name)
	}

	for _, r := range regions {
		if !environmentSet.Has(r.Environment) {
			return perrors.Wrapf(ErrEnvironmentNotFound, "environment %s is not exists", r.Environment)
		}
		if !regionSet.Has(r.Region) {
			return perrors.Wrapf(ErrRegionNotFound, "region %s is not exists", r.Region)
		}
		applicationRegions = append(applicationRegions, &models.ApplicationRegion{
			ApplicationID:   applicationID,
			EnvironmentName: r.Environment,
			RegionName:      r.Region,
			CreatedBy:       currentUser.GetID(),
			UpdatedBy:       currentUser.GetID(),
		})
	}

	return c.mgr.UpsertByApplicationID(ctx, applicationID, applicationRegions)
}
