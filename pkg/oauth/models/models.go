package models

import (
	"time"
)

// type userType uint8

// const (
// 	user  userType = 1
// 	robot userType = 2
// )

type OwnerType uint8

const (
	GroupOwnerType OwnerType = 1
)

type Token struct {
	ID uint `gorm:"primarykey"`

	// grant client info
	ClientID    string `gorm:"column:client_id"`
	RedirectURI string `gorm:"column:redirect_url"`
	State       string `gorm:"column:state"`

	// token basic info
	// Code authorize_code/access_token/refresh-token
	Code      string        `gorm:"column:code"`
	CreatedAt time.Time     `gorm:"column:created_at"`
	ExpiresIn time.Duration `gorm:"column:expires_in"`
	Scope     string        `gorm:"column:scope"`

	UserOrRobotIdentity string `gorm:"column:user_or_robot_identity"`
}
type AppType uint8

const (
	HorizonOAuthAPP AppType = 1
	DirectOAuthAPP  AppType = 2
)

type OauthApp struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"column:name"`
	ClientID    string    `gorm:"column:client_id"`
	RedirectURL string    `gorm:"column:redirect_url"`
	HomeURL     string    `gorm:"column:home_url"`
	Desc        string    `gorm:"column:desc"`
	OwnerType   OwnerType `gorm:"column:owner_type"`
	OwnerID     uint
	AppType     AppType `gorm:"column:app_type"`
}

type OauthClientSecret struct {
	ID           uint      `gorm:"column:id"`
	ClientID     string    `gorm:"column:client_id"`
	ClientSecret string    `gorm:"column:client_secret"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	CreatedBy    uint      `gorm:"column:created_by"`
}
