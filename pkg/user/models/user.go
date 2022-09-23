package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type User struct {
	global.Model

	Name     string
	FullName string
	Email    string
	Phone    string
	Admin    bool
	Banned   bool
}

type IDPUserRelationship struct {
	global.Model

	Sub    string
	IdpID  string
	UserID string
	Name   string
	Email  string
}

func (IDPUserRelationship) TableName() string {
	return "tb_idp_user"
}
