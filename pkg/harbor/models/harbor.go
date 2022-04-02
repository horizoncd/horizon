package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Harbor struct {
	gorm.Model

	Server          string
	Token           string
	PreheatPolicyID int
	DeletedTs       soft_delete.DeletedAt
}
