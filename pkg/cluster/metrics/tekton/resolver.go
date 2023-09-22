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

package tekton

import (
	"sort"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"

	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
)

type WrappedPipelineRun struct {
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

type PrMetadata struct {
	Name      string
	Namespace string
	Pipeline  string
}

type PrResult struct {
	DurationSeconds float64
	Result          string
	StartTime       *metav1.Time
	CompletionTime  *metav1.Time
}

type TrResult struct {
	Name            string
	Pod             string
	Task            string
	StartTime       *metav1.Time
	CompletionTime  *metav1.Time
	DurationSeconds float64
	Result          string
}

type TrResults []*TrResult

func (t TrResults) Len() int {
	return len(t)
}

func (t TrResults) Less(i, j int) bool {
	return t[i].StartTime.Before(t[j].StartTime)
}

func (t TrResults) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type StepResult struct {
	Step            string
	Task            string
	TaskRun         string
	StartTime       *metav1.Time
	CompletionTime  *metav1.Time
	DurationSeconds float64
	Result          string
}

type StepResults []*StepResult

func (s StepResults) Len() int {
	return len(s)
}

func (s StepResults) Less(i, j int) bool {
	return s[i].StartTime.Before(s[j].StartTime)
}

func (s StepResults) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type PipelineResults struct {
	Metadata    *PrMetadata
	PrResult    *PrResult
	TrResults   TrResults
	StepResults StepResults
}

func FormatPipelineResults(pipelineRun *v1beta1.PipelineRun) *PipelineResults {
	if pipelineRun == nil {
		return nil
	}

	wpr := WrappedPipelineRun{
		PipelineRun: pipelineRun,
	}

	trResults, stepResults := wpr.ResolveTrAndStepResults()
	return &PipelineResults{
		Metadata:    wpr.ResolveMetadata(),
		PrResult:    wpr.ResolvePrResult(),
		TrResults:   trResults,
		StepResults: stepResults,
	}
}

const LabelKeyPipeline = "tekton.dev/pipeline"

func (wpr *WrappedPipelineRun) ResolveMetadata() *PrMetadata {
	return &PrMetadata{
		Name:      wpr.PipelineRun.Name,
		Namespace: wpr.PipelineRun.Namespace,
		Pipeline:  wpr.PipelineRun.Labels[LabelKeyPipeline],
	}
}

func (wpr *WrappedPipelineRun) ResolvePrResult() *PrResult {
	r := func() prmodels.PipelineStatus {
		prc := wpr.PipelineRun.Status.GetCondition(apis.ConditionSucceeded)
		if prc == nil {
			return prmodels.StatusUnknown
		}
		switch v1beta1.PipelineRunReason(prc.GetReason()) {
		case v1beta1.PipelineRunReasonSuccessful, v1beta1.PipelineRunReasonCompleted:
			return prmodels.StatusOK
		case v1beta1.PipelineRunReasonFailed, v1beta1.PipelineRunReasonTimedOut:
			return prmodels.StatusFailed
			// When used with tekton pipelines v0.18.1, v1beta1.PipelineRunSpecStatusCancelled need to be checked actually.
			// ref: (1) https://github.com/tektoncd/pipeline/blob/v0.18.1/pkg/reconciler/pipelinerun/cancel.go#L67
			// (2) https://github.com/tektoncd/pipeline/blob/v0.18.1/pkg/reconciler/pipelinerun/pipelinerun.go#L99
		case v1beta1.PipelineRunReasonCancelled, v1beta1.PipelineRunSpecStatusCancelled:
			return prmodels.StatusCancelled
		}
		return prmodels.StatusUnknown
	}()

	return &PrResult{
		DurationSeconds: durationSeconds(
			wpr.PipelineRun.Status.StartTime,
			wpr.PipelineRun.Status.CompletionTime),
		Result:         string(r),
		StartTime:      wpr.PipelineRun.Status.StartTime,
		CompletionTime: wpr.PipelineRun.Status.CompletionTime,
	}
}

func (wpr *WrappedPipelineRun) ResolveTrAndStepResults() (TrResults, StepResults) {
	trResults := make(TrResults, 0)
	stepResults := make(StepResults, 0)

	for trName, trStatus := range wpr.PipelineRun.Status.TaskRuns {
		if trStatus == nil || trStatus.Status == nil {
			continue
		}

		trResults = append(trResults, &TrResult{
			Name: trName,
			Pod:  trStatus.Status.PodName,
			Task: trStatus.PipelineTaskName,
			DurationSeconds: durationSeconds(
				trStatus.Status.StartTime, trStatus.Status.CompletionTime),
			Result:         string(trResult(trStatus)),
			StartTime:      trStatus.Status.StartTime,
			CompletionTime: trStatus.Status.CompletionTime,
		})

		for _, step := range trStatus.Status.Steps {
			stepResult := func() prmodels.PipelineStatus {
				if step.Terminated == nil {
					return prmodels.StatusUnknown
				}
				if step.Terminated.ExitCode == 0 {
					return prmodels.StatusOK
				}
				return prmodels.StatusFailed
			}()
			if stepResult == prmodels.StatusUnknown {
				break
			}
			stepResults = append(stepResults, &StepResult{
				Step:           step.Name,
				Task:           trStatus.PipelineTaskName,
				TaskRun:        trName,
				StartTime:      &step.Terminated.StartedAt,
				CompletionTime: &step.Terminated.FinishedAt,
				DurationSeconds: func() float64 {
					if step.Terminated == nil {
						return -1
					}
					return durationSeconds(
						&step.Terminated.StartedAt, &step.Terminated.FinishedAt)
				}(),
				Result: string(stepResult),
			})
			if stepResult == prmodels.StatusFailed {
				break
			}
		}
	}

	sort.Sort(trResults)
	sort.Sort(stepResults)
	return trResults, stepResults
}

func trResult(trStatus *v1beta1.PipelineRunTaskRunStatus) prmodels.PipelineStatus {
	if trStatus == nil {
		return prmodels.StatusUnknown
	}
	trc := trStatus.Status.GetCondition(apis.ConditionSucceeded)
	switch v1beta1.TaskRunReason(trc.GetReason()) {
	case v1beta1.TaskRunReasonSuccessful:
		return prmodels.StatusOK
	case v1beta1.TaskRunReasonFailed, v1beta1.TaskRunReasonTimedOut:
		return prmodels.StatusFailed
	case v1beta1.TaskRunReasonCancelled:
		return prmodels.StatusCancelled
	}
	return prmodels.StatusUnknown
}

func durationSeconds(beginTime, endTime *metav1.Time) float64 {
	if beginTime == nil || endTime == nil {
		return -1
	}
	return endTime.Time.Sub(beginTime.Time).Seconds()
}
