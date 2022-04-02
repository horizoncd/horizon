package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type TemplateRelease struct {
	gorm.Model
	TemplateName  string
	Name          string
	Description   string
	GitlabProject string
	Recommended   bool
	CreatedBy     uint
	UpdatedBy     uint
	DeletedTs     soft_delete.DeletedAt
}
