package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
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

func ToUser(user *User) *UserBasic {
	if user == nil {
		return nil
	}
	return &UserBasic{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
