package models

import (
	"fmt"

	"github.com/horizoncd/horizon/core/common"
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
