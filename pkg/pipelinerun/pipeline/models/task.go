package models

import "time"

type Task struct {
	ID            uint
	PipelinerunID uint
	Application   string
	Cluster       string
	Region        string
	Pipeline      string
	Task          string
	Result        string
	Duration      uint
	StartedAt     time.Time
	FinishedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
