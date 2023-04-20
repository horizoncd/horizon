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

package cloudevent

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"
)

type WrappedPipelineRun struct {
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

func (wpr *WrappedPipelineRun) IsFinished() bool {
	if wpr.PipelineRun == nil {
		return false
	}
	prc := wpr.PipelineRun.Status.GetCondition(apis.ConditionSucceeded)
	if prc == nil {
		return false
	}
	switch v1beta1.PipelineRunReason(prc.GetReason()) {
	case v1beta1.PipelineRunReasonSuccessful, v1beta1.PipelineRunReasonCompleted,
		v1beta1.PipelineRunReasonFailed, v1beta1.PipelineRunReasonTimedOut,
		v1beta1.PipelineRunReasonCancelled, v1beta1.PipelineRunSpecStatusCancelled:
		return true
	}
	return false
}
