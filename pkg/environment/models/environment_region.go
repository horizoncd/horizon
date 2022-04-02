package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type EnvironmentRegion struct {
	gorm.Model

	EnvironmentName string
	RegionName      string
	Disabled        bool
	CreatedBy       uint
	UpdatedBy       uint
	DeletedTs       soft_delete.DeletedAt
}
