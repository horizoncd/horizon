package cluster

type InternalDeployRequestV2 struct {
	PipelinerunID uint                   `json:"pipelinerunID"`
	Output        map[string]interface{} `json:"output"`
}

type InternalDeployResponseV2 struct {
	PipelinerunID uint   `json:"pipelinerunID"`
	Commit        string `json:"commit"`
}
