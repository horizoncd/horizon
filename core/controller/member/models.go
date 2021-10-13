package member

import (
	"fmt"
	"time"

	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/member/models"
)

type PostMember struct {
	// ResourceType group/application/applicationInstance
	ResourceType string

	// ResourceID group id;application id ...
	ResourceID uint

	// MemberInfo group id / username
	MemberInfo string

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

	// GrantBy user who grant the role
	GrantBy string

	// GrantTime
	GrantTime time.Time
}

func ConvertMember(member *models.Member, sourceInfo string) Member {
	//TODO(tom): change ID to name ,and add the path
	return Member{
		ID:         member.ID,
		MemberType: member.MemberType,
		MemberInfo: member.MemberInfo,
		SourceInfo: sourceInfo,
		Role:       member.Role,
		GrantBy:    member.GrantBy,
		GrantTime:  member.UpdatedAt,
	}
}

func ConvertResourceType(resourceTypeStr string) (models.ResourceType, bool) {
	var convertOk bool = true
	var resourceType models.ResourceType

	switch resourceTypeStr {
	case "group":
		resourceType = models.TypeGroup
	case "application":
		resourceType = models.TypeApplication
	case "applicationInstance":
		resourceType = models.TypeApplicationInstance
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
		MemberInfo:   postMember.MemberInfo,
		GrantBy:      currentUser.GetName(),
	}, nil
}
