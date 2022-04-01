package models

import "gorm.io/gorm"

type Template struct {
	gorm.Model
	Name        string
	Description string
	CreatedBy   uint
	UpdatedBy   uint
	DeletedTs   int64
}
