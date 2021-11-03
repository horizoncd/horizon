package models

import (
	"time"
)

type Pipelinerun struct {
	ID               uint
	ClusterID        uint
	Action           string
	Status           string
	Title            string
	Description      string
	GitURL           string
	GitBranch        string
	GitCommit        string
	ImageURL         string
	LastConfigCommit string
	ConfigCommit     string
	LogBucket        string
	LogObject        string
	StartedAt        *time.Time
	FinishedAt       *time.Time
	RollbackFrom     *uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        uint
}
