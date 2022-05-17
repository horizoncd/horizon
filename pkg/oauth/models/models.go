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

// const (
// 	GroupOwnerType OwnerType = 1
// )

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
	ExpiresIn time.Duration `gorm:"column:expire_in"`
	Scope     string        `gorm:"column:scope"`

	UserOrRobotIdentity string `gorm:"column:scope"`
}

type OauthApp struct {
	ID          uint      `gorm:"column:id"`
	Name        string    `gorm:"column:name"`
	ClientID    string    `gorm:"column:client_id"`
	RedirectURI string    `gorm:"column:redirect_url"`
	HomeURL     string    `gorm:"column:home_url"`
	Desc        string    `gorm:"column:desc"`
	OwnerType   OwnerType `gorm:"column:owner_type"`
	OwnerID     uint
}

type ClientSecret struct {
	ID           uint      `gorm:"column:id"`
	ClientID     string    `gorm:"column:client_id"`
	ClientSecret string    `gorm:"column:client_secret"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	CreateBy     uint      `gorm:"column:create_by"`
}
