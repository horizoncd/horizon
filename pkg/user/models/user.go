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

type UserBasic struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
