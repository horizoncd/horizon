package collector

import (
	"context"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
)

// Object the pipelinerun object to be collected
type Object struct {
	// Metadata meta data
	Metadata *ObjectMeta `json:"metadata"`
	// PipelineRun v1beta1.PipelineRun
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

type (
	ObjectMeta struct {
		Application       string             `json:"application"`
		ApplicationID     uint               `json:"applicationID"`
		Cluster           string             `json:"cluster"`
		ClusterID         uint               `json:"clusterID"`
		Environment       string             `json:"environment"`
		Operator          string             `json:"operator"`
		CreationTimestamp string             `json:"creationTimestamp"`
		PipelineRun       *PipelineRunStatus `json:"pipelineRun"`
	}
	PipelineRunStatus struct {
		StatusMeta `json:",inline"`
		Pipeline   string                   `json:"pipeline"`
		TaskRuns   map[string]TaskRunStatus `json:"taskRuns"`
		TasksOrder []string                 `json:"tasksOrder"`
	}
	TaskRunStatus struct {
		StatusMeta `json:",inline"`
		Pod        string                `json:"pod"`
		Task       string                `json:"task"`
		Steps      map[string]StepStatus `json:"steps"`
		// step的执行顺序
		StepsOrder []string `json:"stepsOrder"`
	}
	StepStatus struct {
		StatusMeta `json:",inline"`
	}
	StatusMeta struct {
		Name            string       `json:"name"`
		Result          string       `json:"result"`
		DurationSeconds float64      `json:"durationSeconds"`
		StartTime       *metav1.Time `json:"startTime"`
		CompletionTime  *metav1.Time `json:"completionTime"`
	}
)

// nolint
//go:generate mockgen -source=$GOFILE -destination=../../../../mock/pkg/cluster/tekton/collector/collector_mock.go
// -package=mock_collector
type Interface interface {
	// Collect log & object for pipelinerun
	Collect(ctx context.Context, pr *v1beta1.PipelineRun, prBusinessData *metrics.PrBusinessData) (*CollectResult, error)

	// GetPipelineRunLog get pipelinerun log from collector
	GetPipelineRunLog(ctx context.Context, logObject string) (_ []byte, err error)

	// GetPipelineRunObject get pipelinerun object from collector
	GetPipelineRunObject(ctx context.Context, object string) (*Object, error)
}

var _ Interface = (*S3Collector)(nil)

const timestampLayout = "20060102150405"

func resolveObjMetadata(pr *v1beta1.PipelineRun, prBusinessData *metrics.PrBusinessData) *ObjectMeta {
	wrappedPr := &metrics.WrappedPipelineRun{
		PipelineRun: pr,
	}
	prMetadata := wrappedPr.ResolveMetadata()
	prResult := wrappedPr.ResolvePrResult()
	trResults, stepResults := wrappedPr.ResolveTrAndStepResults()

	stepMap := make(map[string]map[string]StepStatus)
	stepOrderMap := make(map[string][]string)
	for _, stepResult := range stepResults {
		if _, ok := stepMap[stepResult.TaskRun]; !ok {
			stepMap[stepResult.TaskRun] = make(map[string]StepStatus)
			stepOrderMap[stepResult.TaskRun] = make([]string, 0)
		}
		stepMap[stepResult.TaskRun][stepResult.Step] = StepStatus{
			StatusMeta: StatusMeta{
				Name:            stepResult.Step,
				Result:          stepResult.Result,
				DurationSeconds: stepResult.DurationSeconds,
			},
		}
		stepOrderMap[stepResult.TaskRun] = append(stepOrderMap[stepResult.TaskRun], stepResult.Step)
	}

	taskRuns, tasksOrder := func() (map[string]TaskRunStatus, []string) {
		trMap := make(map[string]TaskRunStatus)
		tasksOrder := make([]string, 0)
		for _, trResult := range trResults {
			trMap[trResult.Name] = TaskRunStatus{
				StatusMeta: StatusMeta{
					Name:            trResult.Name,
					Result:          trResult.Result,
					DurationSeconds: trResult.DurationSeconds,
				},
				Pod:        trResult.Pod,
				Task:       trResult.Task,
				Steps:      stepMap[trResult.Name],
				StepsOrder: stepOrderMap[trResult.Name],
			}
			tasksOrder = append(tasksOrder, trResult.Task)
		}
		return trMap, tasksOrder
	}()

	cstSh, _ := time.LoadLocation("Asia/Shanghai")
	return &ObjectMeta{
		Application:       prBusinessData.Application,
		ApplicationID:     prBusinessData.ApplicationID,
		Cluster:           prBusinessData.Cluster,
		ClusterID:         prBusinessData.ClusterID,
		Environment:       prBusinessData.Environment,
		Operator:          prBusinessData.Operator,
		CreationTimestamp: pr.CreationTimestamp.In(cstSh).Format(timestampLayout),
		PipelineRun: &PipelineRunStatus{
			StatusMeta: StatusMeta{
				Name:            prMetadata.Name,
				Result:          prResult.Result,
				DurationSeconds: prResult.DurationSeconds,
				StartTime:       prResult.StartTime,
				CompletionTime:  prResult.CompletionTime,
			},
			Pipeline:   prMetadata.Pipeline,
			TaskRuns:   taskRuns,
			TasksOrder: tasksOrder,
		},
	}
}
