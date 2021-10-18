package models

import (
	"gorm.io/gorm"
)

type EnvironmentRegion struct {
	gorm.Model

	Env       string
	Region    string
	CreatedBy string
	UpdatedBy string
}
