package models

import "gorm.io/gorm"

type Harbor struct {
	gorm.Model

	Server          string
	Token           string
	PreheatPolicyID int
}
