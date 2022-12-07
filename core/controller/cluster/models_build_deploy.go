package cluster

import "time"

type BuildDeployRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Git         *BuildDeployRequestGit `json:"git"`
}

type BuildDeployRequestGit struct {
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

type BuildDeployResponse struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type GetDiffResponse struct {
	CodeInfo   *CodeInfo `json:"codeInfo"`
	ConfigDiff string    `json:"configDiff"`
}

type CodeInfo struct {
	// deploy branch info
	Branch string `json:"branch,omitempty"`
	// deploy tag info
	Tag string `json:"tag,omitempty"`
	// current branch commit
	CommitID string `json:"commitID"`
	// commit message
	CommitMsg string `json:"commitMsg"`
	// code history link
	Link string `json:"link"`
}

const (
	TokenNameFormat     = "tekton_callback_%s"
	TokenExpiresIn      = time.Hour * 2
	TokenScopeClusterRW = "clusters:read-write"
)
