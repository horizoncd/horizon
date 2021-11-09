package cluster

type BuildDeployRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Git         *BuildDeployRequestGit `json:"git"`
}

type BuildDeployRequestGit struct {
	Branch string `json:"branch"`
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
	Branch string `json:"branch"`
	// current branch commit
	CommitID string `json:"commitID"`
	// commit message
	CommitMsg string `json:"commitMsg"`
	// code history link
	Link string `json:"link"`
}
