package models

import "time"

type ApplicationRegion struct {
	ID              uint
	ApplicationID   uint   `gorm:"uniqueIndex:idx_application_environment"`
	EnvironmentName string `gorm:"uniqueIndex:idx_application_environment"`
	RegionName      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uint
	UpdatedBy       uint
}
