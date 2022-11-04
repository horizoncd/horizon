package cluster

type InternalDeployResponse struct {
	PipelinerunID uint   `json:"pipelinerunID"`
	Commit        string `json:"commit"`
}
