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
	"errors"
	"fmt"
	"time"

	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	clusterservice "github.com/horizoncd/horizon/pkg/cluster/service"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/horizoncd/horizon/pkg/member/models"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/param"
	tmanager "github.com/horizoncd/horizon/pkg/template/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

type UpdateMember struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
}

type PostMember struct {
	// ResourceType group/application/cluster
	ResourceType string `json:"resourceType"`

	// ResourceID group id;application id ...
	ResourceID uint `json:"resourceID"`

	// MemberType user or group
	MemberType models.MemberType `json:"memberType"`

	// MemberNameID group id / userid
	MemberNameID uint `json:"memberNameID"`

	// Role owner/maintainer/develop/...
	Role string `json:"role"`
}

type Member struct {
	// ID the uniq id of the member entry
	ID uint `json:"id"`

	// ResourceName   application/group
	ResourceType models.ResourceType `json:"resourceType"`
	ResourceName string              `json:"resourceName"`
	ResourcePath string              `json:"resourcePath,omitempty"`
	ResourceID   uint                `json:"resourceID"`

	// MemberType user or group
	MemberType models.MemberType `json:"memberType"`

	// MemberName username or groupName
	MemberName string `json:"memberName"`
	// MemberNameID userID or groupID
	MemberNameID uint `json:"memberNameID"`

	// Role the role name that bind
	Role string `json:"role"`
	// GrantedBy id of user who grant the role
	GrantedBy uint `json:"grantedBy"`
	// GrantorName name of user who grant the role
	GrantorName string `json:"grantorName"`
	// GrantTime
	GrantTime time.Time `json:"grantTime"`
}

func CovertPostMember(member *PostMember) memberservice.PostMember {
	return memberservice.PostMember{
		ResourceType: member.ResourceType,
		ResourceID:   member.ResourceID,
		MemberInfo:   member.MemberNameID,
		MemberType:   member.MemberType,
		Role:         member.Role,
	}
}

type ConvertMemberHelp interface {
	ConvertMember(ctx context.Context, member *models.Member) (*Member, error)
	ConvertMembers(ctx context.Context, member []models.Member) ([]Member, error)
}

type converter struct {
	userManager    usermanager.Manager
	groupMgr       groupmanager.Manager
	groupSvc       groupservice.Service
	applicationSvc applicationservice.Service
	clusterSvc     clusterservice.Service
	templateMgr    tmanager.Manager
	releaseMgr     trmanager.Manager
}

func New(param *param.Param) ConvertMemberHelp {
	return &converter{
		userManager:    param.UserMgr,
		groupMgr:       param.GroupMgr,
		groupSvc:       param.GroupSvc,
		applicationSvc: param.ApplicationSvc,
		clusterSvc:     param.ClusterSvc,
		templateMgr:    param.TemplateMgr,
		releaseMgr:     param.TemplateReleaseMgr,
	}
}

func (c *converter) ConvertMember(ctx context.Context, member *models.Member) (_ *Member, err error) {
	// convert userID to userName
	var memberInfo string
	var user *usermodels.User

	if member.MemberType == models.MemberUser {
		user, err = c.userManager.GetUserByID(ctx, member.MemberNameID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
		memberInfo = user.Name
	} else {
		// TODO(tom) covert groupID to GroupName
		return nil, errors.New("group member not support yet")
	}

	return &Member{
		ID:           member.ID,
		MemberType:   member.MemberType,
		MemberName:   memberInfo,
		MemberNameID: member.MemberNameID,
		ResourceType: member.ResourceType,
		ResourceID:   member.ResourceID,
		Role:         member.Role,
		GrantedBy:    member.GrantedBy,
		GrantTime:    member.UpdatedAt,
	}, nil
}
func (c *converter) ConvertMembers(ctx context.Context, members []models.Member) ([]Member, error) {
	var userIDs []uint

	for _, member := range members {
		if member.MemberType != models.MemberUser {
			return nil, errors.New("Only Support User MemberType yet")
		}
		userIDs = append(userIDs, member.MemberNameID, member.GrantedBy)
	}
	users, err := c.userManager.GetUserByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userIDToName := make(map[uint]string)
	for _, userItem := range users {
		userIDToName[userItem.ID] = userItem.Name
	}
	var retMembers []Member
	for _, member := range members {
		var resourceName, resourcePath string
		switch member.ResourceType {
		case models.TypeGroup:
			if !c.groupMgr.IsRootGroup(member.ResourceID) {
				group, err := c.groupSvc.GetChildByID(ctx, member.ResourceID)
				if err != nil {
					return nil, err
				}
				resourceName = group.Name
				resourcePath = group.FullPath
			}
		case models.TypeApplication:
			application, err := c.applicationSvc.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = application.Name
			resourcePath = application.FullPath
		case models.TypeApplicationCluster:
			cluster, err := c.clusterSvc.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = cluster.Name
			resourcePath = cluster.FullPath
		case models.TypeTemplate:
			template, err := c.templateMgr.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = template.Name
			resourcePath = fmt.Sprintf("%d", template.ID)
		case models.TypeTemplateRelease:
			release, err := c.releaseMgr.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			template, err := c.templateMgr.GetByID(ctx, release.Template)
			if err != nil {
				return nil, err
			}
			resourceName = template.Name
			resourcePath = fmt.Sprintf("%d", template.ID)
		default:
			return nil, fmt.Errorf("%s is not support now", member.ResourceType)
		}
		retMembers = append(retMembers, Member{
			ID:           member.ID,
			MemberType:   member.MemberType,
			MemberName:   userIDToName[member.MemberNameID],
			MemberNameID: member.MemberNameID,
			ResourceType: member.ResourceType,
			ResourceID:   member.ResourceID,
			ResourceName: resourceName,
			ResourcePath: resourcePath,
			Role:         member.Role,
			GrantedBy:    member.GrantedBy,
			GrantorName:  userIDToName[member.GrantedBy],
			GrantTime:    member.UpdatedAt,
		})
	}
	return retMembers, nil
}
