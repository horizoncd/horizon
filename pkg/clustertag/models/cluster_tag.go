package models

import "time"

type ClusterTag struct {
	ID        uint   `gorm:"primarykey"`
	ClusterID uint   `gorm:"uniqueIndex:idx_cluster_id_key"`
	Key       string `gorm:"uniqueIndex:idx_cluster_id_key;column:tag_key"`
	Value     string `gorm:"column:tag_value"`
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
}
