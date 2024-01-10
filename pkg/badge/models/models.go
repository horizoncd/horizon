package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type Badge struct {
	global.Model
	ID           uint   `gorm:"primary_key"`
	ResourceID   uint   `gorm:"uniqueIndex:idx_resource_name"`
	ResourceType string `gorm:"uniqueIndex:idx_resource_name"`
	Name         string `gorm:"uniqueIndex:idx_resource_name"`
	SvgLink      string `gorm:"column:svg_link"`
	RedirectLink string `gorm:"column:redirect_link"`
	CreatedBy    uint
	UpdatedBy    uint
}
