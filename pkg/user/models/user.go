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
