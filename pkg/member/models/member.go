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

type Member struct {
	gorm.Model

	// member entry basic info
	// ResourceType group/application/applicationinstance
	ResourceType ResourceType
	// ResourceID  groupID/applicationID/applicationinstanceID
	ResourceID uint

	// role binding info
	// Role: owner/maintainer/...
	Role string
	// MemberType: user/group
	MemberType MemberType `gorm:"column:member_type"`
	//MemberInfo username or groupid
	MemberInfo string `gorm:"column:member_info"`
	GrantBy    string `gorm:"column:grant_by"`
}
