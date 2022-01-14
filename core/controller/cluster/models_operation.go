package cluster

import "g.hz.netease.com/horizon/pkg/cluster/cd"

const ServerlessTemplateName = "serverless"

type PipelinerunIDResponse struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type DeployRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ExecRequest struct {
	PodList []string `json:"podList"`
}

type ExecResponse map[string]ExecResult

type ExecResult struct {
	// Result bool value indicates whether the result is successfully
	Result bool   `json:"result"`
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  error  `json:"error,omitempty"`
}

func ofExecResp(resp map[string]cd.ExecResp) ExecResponse {
	resultMap := make(ExecResponse)
	for k, v := range resp {
		resultMap[k] = ExecResult{
			Result: v.Result,
			Stdout: v.Stdout,
			Stderr: v.Stderr,
			Error:  v.Error,
		}
	}
	return resultMap
}

type RollbackRequest struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type MemcachedSchema struct {
	Enabled bool `json:"enabled"`
}
