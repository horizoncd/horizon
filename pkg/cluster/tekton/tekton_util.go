package tekton

import (
	"context"
	"strings"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"

	"github.com/horizoncd/horizon/pkg/log/wlog"
)

type PipelineRunStatus struct {
	Name        string
	RunningTask *RunningTask
	Status      string
}

type RunningTask struct {
	Name   string
	Status string
}

func GetPipelineRunStatus(ctx context.Context, pr *v1beta1.PipelineRun) *PipelineRunStatus {
	const op = "tekton util: get pipelineRun status"
	defer wlog.Start(ctx, op).StopPrint()

	if pr == nil {
		return nil
	}

	ps := &PipelineRunStatus{
		Name: pr.Name,
	}
	prc := pr.Status.GetCondition(apis.ConditionSucceeded)
	if prc != nil {
		ps.Status = prc.Reason
	}
	runningTask := GetPipelineRunningTask(ctx, pr)
	if strings.Contains(ps.Status, string(v1beta1.PipelineRunReasonCancelled)) {
		runningTask.Status = string(v1beta1.TaskRunReasonCancelled)
	}
	ps.RunningTask = runningTask

	return ps
}

func GetPipelineRunningTask(ctx context.Context, pr *v1beta1.PipelineRun) *RunningTask {
	const op = "tekton util: get pipeline runningTask"
	defer wlog.Start(ctx, op).StopPrint()

	var currentTaskName, currentTaskStatus string

	if pr == nil {
		return nil
	}

	if pr.Status.PipelineSpec == nil || pr.Status.TaskRuns == nil {
		return nil
	}

	taskMap := make(map[string]int)
	tasks := make([]string, 0)
	for index, task := range pr.Status.PipelineSpec.Tasks {
		taskMap[task.Name] = index
		tasks = append(tasks, task.Name)
	}

	taskRunMap := make(map[string]*v1beta1.PipelineRunTaskRunStatus)
	currentTaskIndex := -1
	for _, v := range pr.Status.TaskRuns {
		taskRunMap[v.PipelineTaskName] = v
		index := taskMap[v.PipelineTaskName]
		if index > currentTaskIndex {
			currentTaskIndex = index
		}
	}

	currentTaskName = tasks[currentTaskIndex]
	taskRunStatus := taskRunMap[currentTaskName]
	taskRunCondition := taskRunStatus.Status.GetCondition(apis.ConditionSucceeded)
	if taskRunCondition != nil {
		currentTaskStatus = taskRunCondition.Reason
	} else {
		currentTaskStatus = string(v1beta1.TaskRunReasonRunning)
	}

	return &RunningTask{
		Name:   currentTaskName,
		Status: currentTaskStatus,
	}
}
