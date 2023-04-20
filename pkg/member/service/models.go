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

package service

import (
	"fmt"
	"time"

	"github.com/horizoncd/horizon/core/common"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/member/models"
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
	var convertOk = true
	var resourceType models.ResourceType

	switch resourceTypeStr {
	case common.ResourceGroup:
		resourceType = models.TypeGroup
	case common.ResourceApplication:
		resourceType = models.TypeApplication
	case common.ResourceCluster:
		resourceType = models.TypeApplicationCluster
	case common.ResourceTemplate:
		resourceType = models.TypeTemplate
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
	}, nil
}
