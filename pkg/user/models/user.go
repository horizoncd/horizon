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
	Admin    bool
	Banned   bool
}
