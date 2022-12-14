package global

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedTs soft_delete.DeletedAt
}

type HorizonMetaData struct {
	Application   string
	Cluster       string
	ApplicationID uint
	ClusterID     uint
	Environment   string
	Operator      string
	PipelinerunID uint
	Region        string
	Template      string
	EventID       string
}
