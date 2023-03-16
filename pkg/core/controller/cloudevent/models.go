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
