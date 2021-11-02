package models

import (
	"time"

	"gorm.io/gorm"
)

type Pipelinerun struct {
	gorm.Model

	ClusterID        uint
	Action           string
	Status           string
	Title            string
	Description      string
	CodeBranch       string
	CodeCommit       string
	LastConfigCommit string
	ConfigCommit     string
	LogBucket        string
	LogObject        string
	StartedAt        *time.Time
	FinishedAt       *time.Time
	RollbackFrom     *uint
	CreatedBy        string
}
