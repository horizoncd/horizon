/**
Pr 是 PipelineRun的缩写
Tr 是 TaskRun的缩写
*/

package metrics

import (
	"sort"

	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

type WrappedPipelineRun struct {
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

// PrMetadata pipelineRun的元信息
type PrMetadata struct {
	// pipelineRun的name
	Name string
	// pipelineRun的namespace
	Namespace string
	// pipelineRun对应的pipeline
	Pipeline string
}

// PrBusinessData pipelineRun业务相关参数
type PrBusinessData struct {
	Application   string
	Cluster       string
	ApplicationID uint
	ClusterID     uint
	Environment   string
	Operator      string
	PipelinerunID uint
	Region        string
	Template      string
}

// PrResult pipelineRun结果
type PrResult struct {
	// 花费的时间，单位为秒
	DurationSeconds float64
	// 执行结果
	Result string
	// pipelineRun开始执行的时间，用于排序
	StartTime *metav1.Time
	// pipelineRun执行完成的时间
	CompletionTime *metav1.Time
}

// TrResult taskRun结果
type TrResult struct {
	// taskRun的名称
	Name string
	// 对应的Pod
	Pod string
	// 对应的task的名称
	Task string
	// taskRun开始执行的时间，用于排序
	StartTime *metav1.Time
	// taskRun执行完成的时间
	CompletionTime *metav1.Time
	// 花费的时间，单位为秒
	DurationSeconds float64
	// 执行结果
	Result string
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
	// step名称
	Step string
	// 所属Task名称
	Task string
	// 所属TaskRun名称
	TaskRun string
	// step开始执行的时间
	StartTime *metav1.Time
	// step执行完成的时间
	CompletionTime *metav1.Time
	// 花费的时间，单位为秒
	DurationSeconds float64
	// 执行结果
	Result string
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
	Metadata     *PrMetadata
	BusinessData *PrBusinessData
	PrResult     *PrResult
	TrResults    TrResults
	StepResults  StepResults
}

func FormatPipelineResults(pipelineRun *v1beta1.PipelineRun, prBusiness *PrBusinessData) *PipelineResults {
	if pipelineRun == nil {
		return nil
	}

	wpr := WrappedPipelineRun{
		PipelineRun: pipelineRun,
	}

	trResults, stepResults := wpr.ResolveTrAndStepResults()
	return &PipelineResults{
		Metadata:     wpr.ResolveMetadata(),
		BusinessData: prBusiness,
		PrResult:     wpr.ResolvePrResult(),
		TrResults:    trResults,
		StepResults:  stepResults,
	}
}

const LabelKeyPipeline = "tekton.dev/pipeline"

// ResolveMetadata 解析pipelineRun的元数据
func (wpr *WrappedPipelineRun) ResolveMetadata() *PrMetadata {
	return &PrMetadata{
		Name:      wpr.PipelineRun.Name,
		Namespace: wpr.PipelineRun.Namespace,
		Pipeline:  wpr.PipelineRun.Labels[LabelKeyPipeline],
	}
}

// ResolvePrResult 解析pipelineRun整体的执行结果
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
			// tekton pipelines v0.18.1版本，取消的情况下，
			// 实际用的是v1beta1.PipelineRunSpecStatusCancelled字段，
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

// ResolveTrAndStepResults 解析pipelineRun中包含的taskRun以及各个taskRun中step的执行结果
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
				// ResultUnknown的情况表示pipelineRun取消执行，当前step被取消，此时可以跳过后续step
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
				// 如果一个step失败了，那么后续的step都会跳过执行，故这里跳过后续step
				break
			}
		}
	}

	// 返回的结果按照执行顺序排序
	sort.Sort(trResults)
	sort.Sort(stepResults)
	return trResults, stepResults
}

// trResult 根据 v1beta1.PipelineRunTaskRunStatus 获取 taskRun 的执行结果
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

// durationSeconds 根据起始时间计算以秒为单位的时间差
func durationSeconds(beginTime, endTime *metav1.Time) float64 {
	if beginTime == nil || endTime == nil {
		// -1 代表数据有问题
		return -1
	}
	return endTime.Time.Sub(beginTime.Time).Seconds()
}
