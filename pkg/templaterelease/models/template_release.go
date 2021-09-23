package models

import "gorm.io/gorm"

type TemplateRelease struct {
	gorm.Model
	TemplateName  string
	Name          string
	Description   string
	GitlabName    string
	GitlabProject string
	Recommended   bool
	CreatedBy     string
	UpdatedBy     string
}
