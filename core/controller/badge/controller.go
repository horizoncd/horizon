package badge

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	appmanager "github.com/horizoncd/horizon/pkg/application/manager"
	"github.com/horizoncd/horizon/pkg/badge/manager"
	"github.com/horizoncd/horizon/pkg/badge/models"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	perror "github.com/horizoncd/horizon/pkg/errors"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	CreateBadge(ctx context.Context, resourceType string, resourceID uint, badge *Create) (*Badge, error)
	UpdateBadge(ctx context.Context, id uint, badge *Update) (*Badge, error)
	UpdateBadgeByName(ctx context.Context, resourceType string,
		resourceID uint, name string, badge *Update) (*Badge, error)
	ListBadges(ctx context.Context, resourceType string, resourceID uint) ([]*Badge, error)
	GetBadge(ctx context.Context, id uint) (*Badge, error)
	GetBadgeByName(ctx context.Context, resourceType string, resourceID uint, name string) (*Badge, error)
	DeleteBadge(ctx context.Context, id uint) error
	DeleteBadgeByName(ctx context.Context, resourceType string, resourceID uint, name string) error
}

type controller struct {
	badgeMgr       manager.Manager
	clusterMgr     clustermanager.Manager
	applicationMgr appmanager.Manager
	groupMgr       groupmanager.Manager
}

func NewController(param *param.Param) Controller {
	return &controller{
		badgeMgr:       param.BadgeMgr,
		clusterMgr:     param.ClusterMgr,
		applicationMgr: param.ApplicationMgr,
		groupMgr:       param.GroupMgr,
	}
}

func (c *controller) checkResource(ctx context.Context, resourceType string, resourceID uint) error {
	switch resourceType {
	case common.ResourceApplication:
		if _, err := c.applicationMgr.GetByID(ctx, resourceID); err != nil {
			return err
		}
	case common.ResourceCluster:
		if _, err := c.clusterMgr.GetByID(ctx, resourceID); err != nil {
			return err
		}
	case common.ResourceGroup:
		if _, err := c.groupMgr.GetByID(ctx, resourceID); err != nil {
			return err
		}
	default:
		return perror.Wrapf(herrors.ErrParamInvalid, "invalid resource type: %s", resourceType)
	}
	return nil
}

func (c *controller) CreateBadge(ctx context.Context, resourceType string,
	resourceID uint, badge *Create) (*Badge, error) {
	if err := c.checkResource(ctx, resourceType, resourceID); err != nil {
		return nil, err
	}
	daoBadge := &models.Badge{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Name:         badge.Name,
		SvgLink:      badge.SvgLink,
		RedirectLink: badge.RedirectLink,
	}
	var err error
	daoBadge, err = c.badgeMgr.Create(ctx, daoBadge)
	if err != nil {
		return nil, err
	}
	result := &Badge{}
	result.FromDAO(daoBadge)
	return result, nil
}

func (c *controller) UpdateBadge(ctx context.Context, id uint, badge *Update) (*Badge, error) {
	daoBadge := &models.Badge{
		ID: id,
	}

	if badge.SvgLink != nil {
		daoBadge.SvgLink = *badge.SvgLink
	}
	if badge.RedirectLink != nil {
		daoBadge.RedirectLink = *badge.RedirectLink
	}
	var err error
	daoBadge, err = c.badgeMgr.Update(ctx, daoBadge)
	if err != nil {
		return nil, err
	}
	result := &Badge{}
	result.FromDAO(daoBadge)
	return result, nil
}

func (c *controller) UpdateBadgeByName(ctx context.Context, resourceType string,
	resourceID uint, name string, badge *Update) (*Badge, error) {
	daoBadge := &models.Badge{}
	if badge.SvgLink != nil {
		daoBadge.SvgLink = *badge.SvgLink
	}
	if badge.RedirectLink != nil {
		daoBadge.RedirectLink = *badge.RedirectLink
	}
	var err error
	daoBadge, err = c.badgeMgr.UpdateByName(ctx, resourceType, resourceID, name, daoBadge)
	if err != nil {
		return nil, err
	}
	result := &Badge{}
	result.FromDAO(daoBadge)
	return result, nil
}

func (c *controller) ListBadges(ctx context.Context, resourceType string, resourceID uint) ([]*Badge, error) {
	daoBadges, err := c.badgeMgr.List(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	result := make([]*Badge, len(daoBadges))
	for i, daoBadge := range daoBadges {
		result[i] = &Badge{}
		result[i].FromDAO(daoBadge)
	}
	return result, nil
}

func (c *controller) GetBadge(ctx context.Context, id uint) (*Badge, error) {
	daoBadge, err := c.badgeMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	result := &Badge{}
	result.FromDAO(daoBadge)
	return result, nil
}

func (c *controller) GetBadgeByName(ctx context.Context, resourceType string,
	resourceID uint, name string) (*Badge, error) {
	daoBadge, err := c.badgeMgr.GetByName(ctx, resourceType, resourceID, name)
	if err != nil {
		return nil, err
	}
	result := &Badge{}
	result.FromDAO(daoBadge)
	return result, nil
}

func (c *controller) DeleteBadge(ctx context.Context, id uint) error {
	return c.badgeMgr.Delete(ctx, id)
}

func (c *controller) DeleteBadgeByName(ctx context.Context, resourceType string, resourceID uint, name string) error {
	return c.badgeMgr.DeleteByName(ctx, resourceType, resourceID, name)
}
