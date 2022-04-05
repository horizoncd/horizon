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
	OIDCId   string `gorm:"column:oidc_id"`
	OIDCType string `gorm:"column:oidc_type"`
	Admin    bool
}
