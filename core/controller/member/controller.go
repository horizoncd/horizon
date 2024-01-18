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

package member

import (
	"context"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	eventservice "github.com/horizoncd/horizon/pkg/event/service"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	// CreateMember creates a member of the resource
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

// NewController initializes a new group controller
func NewController(param *param.Param) Controller {
	return &controller{
		memberService: param.MemberService,
		convertHelper: New(param),
		eventSvc:      param.EventSvc,
	}
}

type controller struct {
	memberService memberservice.Service
	convertHelper ConvertMemberHelp
	eventSvc      eventservice.Service
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
	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourceMember, member.ID,
		eventmodels.MemberCreated, nil)
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
	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourceMember, member.ID,
		eventmodels.MemberUpdated, nil)
	return retMember, nil
}

func (c *controller) RemoveMember(ctx context.Context, id uint) error {
	member, err := c.memberService.GetMember(ctx, id)
	if err != nil {
		return err
	}
	err = c.memberService.RemoveMember(ctx, id)
	if err != nil {
		return err
	}
	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourceMember, member.ID,
		eventmodels.MemberDeleted, nil)
	return nil
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
