package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

const (
	UserTypeCommon = iota
	UserTypeRobot
)

type User struct {
	global.Model

	Name     string
	FullName string
	Email    string
	Phone    string
	UserType uint
	Password string
	OidcID   string `gorm:"column:oidc_id"`
	OidcType string `gorm:"column:oidc_type"`
	Admin    bool
	Banned   bool
}

type UserBasic struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
