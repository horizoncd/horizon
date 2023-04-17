package member

import (
	"context"
	"strconv"

	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	// CreateMember create a member of the group
	CreateMember(ctx context.Context, postMember *PostMember) (*Member, error)
	// UpdateMember update a member of the group
	UpdateMember(ctx context.Context, id uint, role string) (*Member, error)
	// RemoveMember leave group or remove a member of the group
	RemoveMember(ctx context.Context, id uint) error
	// ListMember list all the member of the group (and all the member from parent group)
	ListMember(ctx context.Context, resourceType string, resourceID uint) ([]Member, error)
	// GetMemberOfResource get the member of the group by user info in ctx
	GetMemberOfResource(ctx context.Context, resourceType string, resourceID uint) (*Member, error)
}

// NewController initializes a new group controller.
func NewController(param *param.Param) Controller {
	return &controller{
		memberService: param.MemberService,
		convertHelper: New(param),
	}
}

type controller struct {
	memberService memberservice.Service
	convertHelper ConvertMemberHelp
}

func (c *controller) CreateMember(ctx context.Context, postMember *PostMember) (*Member, error) {
	member, err := c.memberService.CreateMember(ctx, CovertPostMember(postMember))
	if err != nil {
		return nil, err
	}
	retMember, err := c.convertHelper.ConvertMember(ctx, member)
	if err != nil {
		return nil, err
	}
	return retMember, nil
}

func (c *controller) UpdateMember(ctx context.Context, id uint, role string) (*Member, error) {
	member, err := c.memberService.UpdateMember(ctx, id, role)
	if err != nil {
		return nil, err
	}
	retMember, err := c.convertHelper.ConvertMember(ctx, member)
	if err != nil {
		return nil, err
	}
	return retMember, nil
}

func (c *controller) RemoveMember(ctx context.Context, id uint) error {
	return c.memberService.RemoveMember(ctx, id)
}

func (c *controller) ListMember(ctx context.Context, resourceType string, id uint) ([]Member, error) {
	members, err := c.memberService.ListMember(ctx, resourceType, id)
	if err != nil {
		return nil, err
	}
	if members == nil || len(members) < 1 {
		return nil, nil
	}
	retMembers, err := c.convertHelper.ConvertMembers(ctx, members)
	if err != nil {
		return nil, err
	}
	return retMembers, nil
}

func (c *controller) GetMemberOfResource(ctx context.Context, resourceType string, resourceID uint) (*Member, error) {
	strID := strconv.FormatUint(uint64(resourceID), 10)
	member, err := c.memberService.GetMemberOfResource(ctx, resourceType, strID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	retMember, err := c.convertHelper.ConvertMember(ctx, member)
	if err != nil {
		return nil, err
	}
	return retMember, nil
}
