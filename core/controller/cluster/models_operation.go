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
	PipelinerunID string `json:"pipelinerunID"`
}
