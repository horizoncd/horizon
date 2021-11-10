package pipelinerun

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
