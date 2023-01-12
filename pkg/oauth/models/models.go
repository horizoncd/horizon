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
	Desc        string    `gorm:"column:description"`
	OwnerType   OwnerType `gorm:"column:owner_type"`
	OwnerID     uint      `gorm:"column:owner_id"`
	AppType     AppType   `gorm:"column:app_type"`

	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy uint      `gorm:"column:created_by"`

	UpdatedAt time.Time `gorm:"column:updated_at"`
	UpdatedBy uint      `gorm:"column:updated_by"`
}

func (a *OauthApp) IsGroupOwnerType() bool {
	return a.OwnerType == GroupOwnerType
}

type OauthClientSecret struct {
	ID           uint      `gorm:"column:id" json:"id"`
	ClientID     string    `gorm:"column:client_id" json:"clientID"`
	ClientSecret string    `gorm:"column:client_secret" json:"clientSecret"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	CreatedBy    uint      `gorm:"column:created_by" json:"createdBy"`
}
