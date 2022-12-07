package models

import (
	"time"
)

type Token struct {
	ID uint `gorm:"primarykey"`

	// grant client info
	ClientID    string `gorm:"column:client_id"`
	RedirectURI string `gorm:"column:redirect_uri"`
	State       string `gorm:"column:state"`

	// token basic info
	Name string `gorm:"column:name"`
	// Code authorize_code/access_token/refresh_token
	Code      string        `gorm:"column:code"`
	CreatedAt time.Time     `gorm:"column:created_at"`
	CreatedBy uint          `gorm:"column:created_by"`
	ExpiresIn time.Duration `gorm:"column:expires_in"`
	Scope     string        `gorm:"column:scope"`

	UserID uint `gorm:"column:user_id"`
}
