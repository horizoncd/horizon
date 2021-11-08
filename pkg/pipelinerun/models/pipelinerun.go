package models

import (
	"time"
)

type Pipelinerun struct {
	// ID pipelinerun id
	ID uint
	// ClusterID cluster id which this pipelinerun belongs to
	ClusterID uint
	// Action type, which can be builddeploy, deploy, restart, rollback
	Action string
	// Status of this pipelinerun, which can be created, ok, failed, cancelled, unknown
	Status string
	// Title of this pipelinerun
	Title string
	// Description of this pipelinerun
	Description string
	// GitURL the git url this pipelinerun to build with, can be empty when action is not builddeploy
	GitURL string
	// GitBranch the git branch this pipelinerun to build with, can be empty when action is not builddeploy
	GitBranch string
	// GitCommit the git commit this pipelinerun to build with, can be empty when action is not builddeploy
	GitCommit string
	// ImageURL image url of this pipelinerun to build image
	ImageURL string
	// the two commit used to compare the config difference of this pipelinerun
	// LastConfigCommit config commit in master branch of this pipelinerun, can be empty when action is restart
	LastConfigCommit string
	// ConfigCommit config commit of this pipelinerun
	ConfigCommit string
	// LogBucket pipelinerun log s3 bucket
	LogBucket string
	// LogObject pipelinerun log s3 object
	LogObject string
	// StartedAt start time of this pipelinerun
	StartedAt *time.Time
	// FinishedAt finish time of this pipelinerun
	FinishedAt *time.Time
	// RollbackFrom which pipelinerun this pipelinerun rollback from
	RollbackFrom *uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uint
}
