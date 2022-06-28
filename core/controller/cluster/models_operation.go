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
	OperationResult
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}

func ofExecResp(resp map[string]cd.ExecResp) ExecResponse {
	resultMap := make(ExecResponse)
	for k, v := range resp {
		execResult := ExecResult{
			Stdout: v.Stdout,
			Stderr: v.Stderr,
		}
		execResult.Result = v.Result
		execResult.Error = v.Error
		resultMap[k] = execResult
	}
	return resultMap
}

type RollbackRequest struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type BatchResponse map[string]OperationResult
type OperationResult struct {
	// Result bool value indicates whether the result is successfully
	Result   bool   `json:"result"`
	Error    error  `json:"error,omitempty"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

func ofBatchResp(resp map[string]cd.OperationResult) BatchResponse {
	resultMap := make(BatchResponse)
	for k, v := range resp {
		opResult := OperationResult{}
		opResult.Result = v.Result
		if v.Error != nil {
			opResult.ErrorMsg = v.Error.Error()
		}
		resultMap[k] = opResult
	}
	return resultMap
}
