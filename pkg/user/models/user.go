package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string
	Email    string
	Phone    string
	OIDCId   string
	OIDCType string
}
