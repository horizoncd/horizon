package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name            string `gorm:"index:idx_name_parent_id,unique"`
	FullName        string
	Path            string `gorm:"index:idx_path,unique"`
	VisibilityLevel string
	Description     string
	ParentId        *uint
}
