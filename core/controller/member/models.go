package member

import (
	"context"
	"errors"
	"fmt"
	"time"

	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	clusterservice "g.hz.netease.com/horizon/pkg/cluster/service"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

var (
	// Converter us tge global member converter
	Converter = New()
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
	userManager usermanager.Manager
}

func New() ConvertMemberHelp {
	return &converter{
		userManager: usermanager.Mgr,
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
			group, err := groupservice.Svc.GetChildByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = group.Name
			resourcePath = group.FullPath
		case models.TypeApplication:
			application, err := applicationservice.Svc.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = application.Name
			resourcePath = application.FullPath
		case models.TypeApplicationCluster:
			cluster, err := clusterservice.Svc.GetByID(ctx, member.ResourceID)
			if err != nil {
				return nil, err
			}
			resourceName = cluster.Name
			resourcePath = cluster.FullPath
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
