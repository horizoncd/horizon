package models

import "time"

type Pipeline struct {
	ID            uint
	PipelinerunID uint
	Application   string
	Cluster       string
	Environment   string
	Pipeline      string
	Result        string
	Duration      uint
	StartedAt     time.Time
	FinishedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StepSLO struct {
	Step     string
	Result   string
	Duration uint
}

type TaskSLO struct {
	Task     string
	Result   string
	Duration uint
	Steps    map[string]*StepSLO
}

type PipelineSLO struct {
	Pipeline string
	Result   string
	Duration uint
	Tasks    map[string]*TaskSLO
}
