package service

import (
	"fmt"
	"time"

	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member/models"
)

type PostMember struct {
	// ResourceType group/application/cluster
	ResourceType string

	// ResourceID group id;application id ...
	ResourceID uint

	// MemberInfo group id / username
	MemberInfo uint

	// MemberType user or group
	MemberType models.MemberType

	// Role owner/maintainer/develop/...
	Role string
}

type Member struct {
	// ID the uniq id of the member entry
	ID uint

	// MemberInfo username or groupID
	MemberInfo string

	// MemberName
	MemberName string

	// MemberPath the path of the member
	MemberPath string

	// MemberType user or group
	MemberType models.MemberType

	// SourceInfo direct or from application/group
	SourceInfo string

	// Role the role name that bind
	Role string

	// GrantedBy user who grant the role
	GrantedBy string

	// GrantTime
	GrantTime time.Time
}

func ConvertResourceType(resourceTypeStr string) (models.ResourceType, bool) {
	var convertOk bool = true
	var resourceType models.ResourceType

	switch resourceTypeStr {
	case models.TypeGroupStr:
		resourceType = models.TypeGroup
	case models.TypeApplicationStr:
		resourceType = models.TypeApplication
	case models.TypeApplicationClusterStr:
		resourceType = models.TypeApplicationCluster
	default:
		convertOk = false
	}
	return resourceType, convertOk
}

func ConvertPostMemberToMember(postMember PostMember, currentUser userauth.User) (*models.Member, error) {
	resourceType, err := ConvertResourceType(postMember.ResourceType)
	if !err {
		return nil, fmt.Errorf("cannot convert ResourceType{%v}",
			postMember.ResourceType)
	}
	return &models.Member{
		ResourceType: resourceType,
		ResourceID:   postMember.ResourceID,
		Role:         postMember.Role,
		MemberType:   postMember.MemberType,
		MemberNameID: postMember.MemberInfo,
		GrantedBy:    currentUser.GetID(),
		CreatedBy:    currentUser.GetID(),
	}, nil
}
