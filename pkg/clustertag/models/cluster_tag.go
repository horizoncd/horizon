package models

import "time"

type ClusterTag struct {
	ID        uint   `gorm:"primarykey"`
	ClusterID uint   `gorm:"uniqueIndex:idx_cluster_id_key"`
	Key       string `gorm:"uniqueIndex:idx_cluster_id_key"`
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
}
