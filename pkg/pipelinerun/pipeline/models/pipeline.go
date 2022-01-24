package models

import "time"

type Pipeline struct {
	ID            uint
	PipelinerunID uint
	Application   string
	Cluster       string
	RegionID      uint
	Pipeline      string
	Result        string
	Duration      uint
	StartedAt     time.Time
	FinishedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
