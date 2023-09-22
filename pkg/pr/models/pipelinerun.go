// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type PipelineStatus string

const (
	StatusCreated PipelineStatus = "created"
	StatusRunning PipelineStatus = "running"
	StatusPending PipelineStatus = "pending"
	// StatusReady means the pipeline is ready to be executed
	StatusReady     PipelineStatus = "ready"
	StatusCommitted PipelineStatus = "committed"
	StatusMerged    PipelineStatus = "merged"
	StatusDeployed  PipelineStatus = "deployed"
	StatusOK        PipelineStatus = "ok"
	StatusFailed    PipelineStatus = "failed"
	StatusCancelled PipelineStatus = "cancelled"
	StatusUnknown   PipelineStatus = "unknown"
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
	// GitRef the git reference this pipelinerun to build with, can be empty when action is not builddeploy
	GitRef string
	// GitRefType the git reference type this pipelinerun to build with, can be empty when action is not builddeploy
	GitRefType string
	// GitCommit the git commit this pipelinerun to build with, can be empty when action is not builddeploy
	GitCommit string
	// ImageURL image url of this pipelinerun to build or deploy image
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
	// CIEventID event id returned from tekton-trigger EventListener
	CIEventID string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uint
}

type PipelineBasic struct {
	// ID pipelinerun id
	ID    uint   `json:"id"`
	Title string `json:"title"`
	// Description of this pipelinerun
	Description string `json:"description"`

	// Action type, which can be builddeploy, deploy, restart, rollback
	Action string `json:"action"`
	// Status of this pipelinerun, which can be created, ok, failed, cancelled, unknown
	Status string `json:"status"`
	// Title of this pipelinerun

	// GitURL the git url this pipelinerun to build with, can be empty when action is not builddeploy
	GitURL string `json:"gitURL"`
	// GitBranch the git branch this pipelinerun to build with, can be empty when action is not builddeploy
	GitBranch string `json:"gitBranch,omitempty"`
	// GitTag the git tag this pipelinerun to build with, can be empty when action is not builddeploy
	GitTag string `json:"gitTag,omitempty"`
	// GitCommit the git commit this pipelinerun to build with, can be empty when action is not builddeploy
	GitCommit string `json:"gitCommit"`
	// ImageURL image url of this pipelinerun to build image
	ImageURL string `json:"imageURL"`

	// LastConfigCommit config commit in master branch of this pipelinerun, can be empty when action is restart
	LastConfigCommit string `json:"lastConfigCommit"`
	// ConfigCommit config commit of this pipelinerun
	ConfigCommit string `json:"configCommit"`
	// CreatedAt create time of this pipelinerun
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt update time of this pipelinerun
	UpdatedAt time.Time `json:"updatedAt"`
	// StartedAt start time of this pipelinerun
	StartedAt *time.Time `json:"startedAt"`
	// FinishedAt finish time of this pipelinerun
	FinishedAt *time.Time `json:"finishedAt"`
	// CanRollback can this pipelinerun be rollback, default is false
	CanRollback bool `json:"canRollback"`
	// createInfo
	CreatedBy UserInfo `json:"createdBy"`
}

type UserInfo struct {
	UserID   uint   `json:"userID"`
	UserName string `json:"userName"`
}

type Result struct {
	S3Bucket   string
	LogObject  string
	PrObject   string
	Result     string
	StartedAt  *time.Time
	FinishedAt *time.Time
}
