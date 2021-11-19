package models

import (
	"time"
)

const (
	ActionBuildDeploy = "builddeploy"
	ActionDeploy      = "deploy"
	ActionRestart     = "restart"
	ActionRollback    = "rollback"
)

const (
	ResultCreated   = "created"
	ResultOK        = "ok"
	ResultFailed    = "failed"
	ResultCancelled = "cancelled"
	ResultUnknown   = "unknown"
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
	// S3Bucket pipelinerun log and object s3 bucket
	S3Bucket string `gorm:"column:s3_bucket"`
	// LogObject pipelinerun's log s3 object
	LogObject string
	// PrObject pipelinerun s3 object
	PrObject string `gorm:"column:pr_object"`
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

type Result struct {
	S3Bucket   string
	LogObject  string
	PrObject   string
	Result     string
	StartedAt  *time.Time
	FinishedAt *time.Time
}