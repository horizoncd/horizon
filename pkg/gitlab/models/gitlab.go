package models

import "gorm.io/gorm"

type Gitlab struct {
	gorm.Model
	Name      string
	URL       string
	Token     string
	CreatedBy uint
	UpdatedBy uint
}
