// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"github.com/horizoncd/horizon/pkg/cd"
)

const ServerlessTemplateName = "serverless"

type PipelinerunIDResponse struct {
	PipelinerunID uint `json:"pipelinerunID"`
}

type DeployRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageTag    string `json:"imageTag"`
}

type ExecuteActionRequest struct {
	Action   string `json:"action"`
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

type ExecRequest struct {
	Commands []string `json:"commands"`
	PodList  []string `json:"podList"`
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
