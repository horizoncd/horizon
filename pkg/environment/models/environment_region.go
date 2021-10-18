package models

import (
	"gorm.io/gorm"
)

type EnvironmentRegion struct {
	gorm.Model

	EnvironmentName string
	RegionName      string
	CreatedBy       string
	UpdatedBy       string
}
