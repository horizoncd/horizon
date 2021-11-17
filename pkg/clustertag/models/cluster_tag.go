package models

import "time"

type ClusterTag struct {
	ID        uint `gorm:"primarykey"`
	ClusterID uint
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
}
