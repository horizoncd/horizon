package models

import "gorm.io/gorm"

type ResourceType string

const (
	TypeGroupStr string = "group"
	// TypeGroup represent the group member entry.
	TypeGroup ResourceType = (ResourceType)(TypeGroupStr)

	TypeApplicationStr string = "application"
	// TypeApplication represent the application  member entry.
	TypeApplication ResourceType = (ResourceType)(TypeApplicationStr)

	TypeApplicationInstanceStr string = "applicationinstance"
	// TypeApplicationInstance represent the application instance member entry
	TypeApplicationInstance ResourceType = (ResourceType)(TypeApplicationInstanceStr)
)

type MemberType uint8

const (
	// MemberUser represent the user binding.
	MemberUser MemberType = iota
	// MemberGroup represent the group binding.
	MemberGroup
)

const (
	Owner     string = "Owner"
	Maitainer string = "Maitainer"
)

type Member struct {
	gorm.Model

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
