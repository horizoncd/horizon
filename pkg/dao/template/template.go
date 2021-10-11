package template

import "gorm.io/gorm"

type Template struct {
	gorm.Model
	Name        string
	Description string
	CreatedBy   string
	UpdatedBy   string
}
