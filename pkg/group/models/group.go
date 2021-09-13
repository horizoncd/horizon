package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name            string
	FullName        string
	Path            string
	VisibilityLevel string
	Description     string
	ParentID        *uint
}
