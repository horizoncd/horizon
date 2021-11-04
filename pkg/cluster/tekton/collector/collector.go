// collector 负责上传pipelineRun相关资源（crd定义、日志等）到远端存储
package collector

import (
	"context"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/metrics"
)

// Object 需要收集起来的pipelineRun元数据
type Object struct {
	// 元数据
	Metadata *ObjectMeta `json:"metadata"`
	// pipelineRun
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

type (
	ObjectMeta struct {
		Application       string             `json:"application"`
		Cluster           string             `json:"cluster"`
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
		Name            string  `json:"name"`
		Result          string  `json:"result"`
		DurationSeconds float64 `json:"durationSeconds"`
	}
)

type Interface interface {
	// Collect 收集pipelineRun的资源对象 & 日志
	Collect(ctx context.Context, pr *v1beta1.PipelineRun) error

	// Delete 删除应用、集群对应的所有内容。释放集群时需要
	Delete(ctx context.Context, application, cluster string) error

	// GetLatestPipelineRunLog 根据应用、集群获取该应用集群最近一次的构建日志
	GetLatestPipelineRunLog(ctx context.Context, application, cluster string) ([]byte, error)

	// GetLatestPipelineRunObject 根据应用、集群获取该应用集群最近一次的资源对象
	GetLatestPipelineRunObject(ctx context.Context, application, cluster string) (*Object, error)
}

var _ Interface = (*S3Collector)(nil)

const CollectorTimeout = 15 * time.Second

const timestampLayout = "20060102150405"

func resolveObjMetadata(pr *v1beta1.PipelineRun) *ObjectMeta {
	wrappedPr := &metrics.WrappedPipelineRun{
		PipelineRun: pr,
	}
	prMetadata := wrappedPr.ResolveMetadata()
	prBusinessData := wrappedPr.ResolveBusinessData()
	prResult := wrappedPr.ResolvePrResult()
	trResults, stepResults := wrappedPr.ResolveTrAndStepResults()

	stepMap := make(map[string]map[string]StepStatus)
	stepOrderMap := make(map[string][]string)
	for _, stepResult := range stepResults {
		if _, ok := stepMap[stepResult.TaskRun]; !ok {
			stepMap[stepResult.TaskRun] = make(map[string]StepStatus)
			stepOrderMap[stepResult.TaskRun] = make([]string, 0)
		}
		stepMap[stepResult.TaskRun][stepResult.Name] = StepStatus{
			StatusMeta: StatusMeta{
				Name:            stepResult.Name,
				Result:          stepResult.Result.String(),
				DurationSeconds: stepResult.DurationSeconds,
			},
		}
		stepOrderMap[stepResult.TaskRun] = append(stepOrderMap[stepResult.TaskRun], stepResult.Name)
	}

	taskRuns, tasksOrder := func() (map[string]TaskRunStatus, []string) {
		trMap := make(map[string]TaskRunStatus)
		tasksOrder := make([]string, 0)
		for _, trResult := range trResults {
			trMap[trResult.Name] = TaskRunStatus{
				StatusMeta: StatusMeta{
					Name:            trResult.Name,
					Result:          trResult.Result.String(),
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
		Cluster:           prBusinessData.Cluster,
		Environment:       prBusinessData.Environment,
		Operator:          prBusinessData.Operator,
		CreationTimestamp: pr.CreationTimestamp.In(cstSh).Format(timestampLayout),
		PipelineRun: &PipelineRunStatus{
			StatusMeta: StatusMeta{
				Name:            prMetadata.Name,
				Result:          prResult.Result.String(),
				DurationSeconds: prResult.DurationSeconds,
			},
			Pipeline:   prMetadata.Pipeline,
			TaskRuns:   taskRuns,
			TasksOrder: tasksOrder,
		},
	}
}
