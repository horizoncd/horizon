package models

import (
	"gorm.io/gorm"
)

type EnvironmentRegion struct {
	gorm.Model

	EnvironmentName string
	RegionName      string
	Disabled        bool
	CreatedBy       uint
	UpdatedBy       uint
	DeletedTs       int64
}
