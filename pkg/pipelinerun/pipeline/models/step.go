package models

import "time"

type Step struct {
	ID            uint
	PipelinerunID uint
	Application   string
	Cluster       string
	RegionID      uint
	Pipeline      string
	Task          string
	Step          string
	Result        string
	Duration      uint
	StartedAt     time.Time
	FinishedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
