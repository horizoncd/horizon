package cluster

type InternalDeployRequest struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type InternalDeployResponse struct {
	PipelinerunID uint   `json:"pipelinerunID"`
	Commit        string `json:"commit"`
}
