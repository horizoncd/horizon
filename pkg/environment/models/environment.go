package models

import "gorm.io/gorm"

type Environment struct {
	gorm.Model

	Name        string
	DisplayName string
	CreatedBy   uint
	UpdatedBy   uint
}
