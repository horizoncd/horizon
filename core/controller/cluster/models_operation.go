package cluster

type PipelinerunIDResponse struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type DeployRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
