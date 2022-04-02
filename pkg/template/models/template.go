package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Template struct {
	gorm.Model
	Name        string
	Description string
	CreatedBy   uint
	UpdatedBy   uint
	DeletedTs   soft_delete.DeletedAt
}
