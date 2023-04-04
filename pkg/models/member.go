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

package models

import (
	"fmt"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/server/global"
)

type ResourceType string

const (
	// TypeGroup represent the group member entry.
	TypeGroup = ResourceType(common.ResourceGroup)

	// TypeApplication represent the application  member entry.
	TypeApplication = ResourceType(common.ResourceApplication)

	// TypeApplicationCluster represent the application instance member entry
	TypeApplicationCluster = ResourceType(common.ResourceCluster)
	// TypePipelinerunStr currently pipelineruns do not have direct member info, will
	// use the pipeline's cluster's member info
	TypePipelinerun = ResourceType(common.ResourcePipelinerun)

	// TypeOauthAppsStr urrently oauthapp do not have direct member info, will
	// use the oauthapp's groups member info
	TypeOauthApps = ResourceType(common.ResourceOauthApps)

	TypeTemplate = ResourceType(common.ResourceTemplate)

	TypeTemplateRelease = ResourceType(common.ResourceTemplateRelease)

	TypeRegion = ResourceType(common.ResourceRegion)
)

type MemberType uint8

const (
	// MemberUser represent the user binding.
	MemberUser MemberType = iota
	// MemberGroup represent the group binding.
	MemberGroup
)

type Member struct {
	global.Model

	// member entry basic info
	// ResourceType group/application/cluster
	ResourceType ResourceType `gorm:"column:resource_type"`
	// ResourceID  groupID/applicationID/applicationinstanceID
	ResourceID uint `gorm:"colum:resource_id"`

	// role binding info
	// Role: owner/maintainer/...
	Role string
	// MemberType: user/group
	MemberType MemberType `gorm:"column:member_type"`
	// userID or groupID
	MemberNameID uint `gorm:"column:membername_id"`

	// TODO(tom): change go user
	GrantedBy uint `gorm:"column:granted_by"`
	CreatedBy uint `gorm:"column:created_by"`
}

func (m *Member) BaseInfo() string {
	return fmt.Sprintf("resource(%s/%d)-memberInfo(%d/%d)-ruleID(%d)",
		m.ResourceType, m.ResourceID, m.MemberType, m.MemberNameID, m.ID)
}

type PostMember struct {
	// ResourceType group/application/cluster
	ResourceType string

	// ResourceID group id;application id ...
	ResourceID uint

	// MemberInfo group id / username
	MemberInfo uint

	// MemberType user or group
	MemberType MemberType

	// Role owner/maintainer/develop/...
	Role string
}

func ConvertResourceType(resourceTypeStr string) (ResourceType, bool) {
	var convertOk = true
	var resourceType ResourceType

	switch resourceTypeStr {
	case common.ResourceGroup:
		resourceType = TypeGroup
	case common.ResourceApplication:
		resourceType = TypeApplication
	case common.ResourceCluster:
		resourceType = TypeApplicationCluster
	case common.ResourceTemplate:
		resourceType = TypeTemplate
	default:
		convertOk = false
	}
	return resourceType, convertOk
}

func ConvertPostMemberToMember(postMember PostMember, currentUser user.User) (*Member, error) {
	resourceType, err := ConvertResourceType(postMember.ResourceType)
	if !err {
		return nil, fmt.Errorf("cannot convert ResourceType{%v}",
			postMember.ResourceType)
	}
	return &Member{
		ResourceType: resourceType,
		ResourceID:   postMember.ResourceID,
		Role:         postMember.Role,
		MemberType:   postMember.MemberType,
		MemberNameID: postMember.MemberInfo,
		GrantedBy:    currentUser.GetID(),
	}, nil
}
