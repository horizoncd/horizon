package pipelinerun

import (
	"time"
)

type GetDiffResponse struct {
	CodeInfo   *CodeInfo   `json:"codeInfo"`
	ConfigDiff *ConfigDiff `json:"configDiff"`
}

type CodeInfo struct {
	// deploy branch info
	Branch string `json:"branch"`
	// branch commit
	CommitID string `json:"commitID"`
	// commit message
	CommitMsg string `json:"commitMsg"`
	// code history link
	Link string `json:"link"`
}

type ConfigDiff struct {
	From string `json:"from"`
	To   string `json:"to"`
	Diff string `json:"diff"`
}

type PipelineBasic struct {
	// ID pipelinerun id
	ID    uint
	Title string
	// Description of this pipelinerun
	Description string

	// Action type, which can be builddeploy, deploy, restart, rollback
	Action string
	// Status of this pipelinerun, which can be created, ok, failed, cancelled, unknown
	Status string
	// Title of this pipelinerun

	// GitURL the git url this pipelinerun to build with, can be empty when action is not builddeploy
	GitURL string
	// GitBranch the git branch this pipelinerun to build with, can be empty when action is not builddeploy
	GitBranch string
	// GitCommit the git commit this pipelinerun to build with, can be empty when action is not builddeploy
	GitCommit string
	// ImageURL image url of this pipelinerun to build image
	ImageURL string

	// LastConfigCommit config commit in master branch of this pipelinerun, can be empty when action is restart
	LastConfigCommit string
	// ConfigCommit config commit of this pipelinerun
	ConfigCommit string

	StartedAt *time.Time
	// FinishedAt finish time of this pipelinerun
	FinishedAt *time.Time
	// createInfo
	CreatedBy UserInfo
}

type UserInfo struct {
	UserID   uint
	UserName string
}

type ListPipelineResponse struct {
	TotalCount int             `json:"totalCount"`
	Pipelines  []PipelineBasic `json:"pipelines"`
}
