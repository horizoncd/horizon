package models

import "gorm.io/gorm"

type Environment struct {
	gorm.Model

	Env       string
	Name      string
	CreatedBy string
	UpdatedBy string
}
