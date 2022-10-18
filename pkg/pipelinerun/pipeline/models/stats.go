package models

import "time"

type StepStats struct {
	Step     string `json:"step"`
	Result   string `json:"result"`
	Duration uint   `json:"duration"`
}

type TaskStats struct {
	Task     string       `json:"task"`
	Result   string       `json:"result"`
	Duration uint         `json:"duration"`
	Steps    []*StepStats `json:"steps"`
}

type PipelineStats struct {
	PipelinerunID uint         `json:"pipelinerunID"`
	Application   string       `json:"application"`
	Cluster       string       `json:"cluster"`
	Pipeline      string       `json:"pipeline"`
	Result        string       `json:"result"`
	Duration      uint         `json:"duration"`
	Tasks         []*TaskStats `json:"tasks"`
	StartedAt     time.Time    `json:"startedAt"`
}
