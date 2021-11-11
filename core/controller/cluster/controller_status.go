package cluster

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/tekton"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const (
	_taskNone          = "none"
	_taskDeploy        = "deploy"
	_taskBuild         = "build"
	_taskStatusPending = "Pending"

	_notFound = "NotFound"
)

func (c *controller) GetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error) {
	const op = "cluster controller: get cluster status"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// get latest builddeploy action pipelinerun
	pr, err := c.pipelinerunMgr.GetLatestByClusterIDAndAction(ctx, clusterID, prmodels.ActionBuildDeploy)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// if latest builddeploy pr is not exists
	if pr == nil {
		return &GetClusterStatusResponse{
			RunningTask: &RunningTask{
				Task: "none",
			},
			ClusterStatus: map[string]string{
				"status": _notFound,
			},
		}, nil
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var latestPr *v1beta1.PipelineRun
	if pr.PrObject == "" {
		tektonClient, err := c.tektonFty.GetTekton(er.EnvironmentName)
		if err != nil {
			return nil, errors.E(op, err)
		}
		latestPr, err = tektonClient.GetPipelineRunByID(ctx, cluster.Name, cluster.ID, pr.ID)
		if err != nil {
			return nil, errors.E(op, err)
		}
	} else {
		tektonCollector, err := c.tektonFty.GetTektonCollector(er.EnvironmentName)
		if err != nil {
			return nil, errors.E(op, err)
		}
		obj, err := tektonCollector.GetPipelineRunObject(ctx, pr.PrObject)
		if err != nil {
			return nil, errors.E(op, err)
		}
		latestPr = obj.PipelineRun
	}

	resp := &GetClusterStatusResponse{}
	resp.RunningTask = c.getRunningTask(ctx, latestPr)
	resp.RunningTask.PipelinerunID = pr.ID

	clusterState, err := c.cd.GetClusterState(ctx, &cd.GetClusterStateParams{
		Environment:  er.EnvironmentName,
		Cluster:      cluster.Name,
		Namespace:    fmt.Sprintf("%v-%v", er.EnvironmentName, application.GroupID),
		RegionEntity: regionEntity,
	})
	if err != nil {
		if errors.Status(err) == http.StatusNotFound {
			resp.ClusterStatus = map[string]string{
				"status": _notFound,
			}
		} else {
			return nil, errors.E(op, err)
		}
	} else {
		resp.ClusterStatus = clusterState
		if resp.RunningTask.Task == _taskNone && clusterState.Status != "" {
			// task为none的情况下，判断是否单独在发布（重启、回滚等）
			// 如果是的话，把runningTask.task设置为deploy，并设置runningTask.taskStatus对应的状态，
			// 这个逻辑是因为overmind想统一通过runningTask判断是否发布完成
			if clusterState.Status == health.HealthStatusDegraded {
				// Degraded 认为是发布失败Failed
				resp.RunningTask.Task = _taskDeploy
				resp.RunningTask.TaskStatus = string(v1beta1.TaskRunReasonFailed)
			} else if clusterState.Status != health.HealthStatusHealthy {
				// 除了Degraded和Healthy之外的状态（Progressing、Suspended）认为是发布中Running
				resp.RunningTask.Task = _taskDeploy
				resp.RunningTask.TaskStatus = string(v1beta1.TaskRunReasonRunning)
			}
		}
	}

	return resp, nil
}

// getRunningTask 获取pipelineRun当前最近一次处于执行中的Task。如果最近一次执行的pipelineRun是成功状态，
// 则返回noneRunningTask
func (c *controller) getRunningTask(ctx context.Context, pr *v1beta1.PipelineRun) *RunningTask {
	noneRunningTask := &RunningTask{
		Task: _taskNone,
	}
	if pr == nil {
		return noneRunningTask
	}

	prs := tekton.GetPipelineRunStatus(ctx, pr)

	isPrFinished := func(pr *tekton.PipelineRunStatus) bool {
		status := prs.Status
		if status == string(v1beta1.PipelineRunReasonSuccessful) ||
			status == string(v1beta1.PipelineRunReasonCompleted) {
			return true
		}
		return false
	}

	if isPrFinished(prs) {
		// 流水线处于完成状态，则返回noneRunningTask。表示当前没有在执行中的Task
		return noneRunningTask
	}

	runningTask := prs.RunningTask
	if runningTask == nil {
		// 如果流水线还未开始执行，则返回build是Pending状态
		return &RunningTask{
			Task:       _taskBuild,
			TaskStatus: _taskStatusPending,
		}
	}
	taskStatus := strings.TrimPrefix(runningTask.Status, "TaskRun")
	// 如果是Timeout，则认为是Failed
	if taskStatus == strings.TrimPrefix(string(v1beta1.TaskRunReasonTimedOut), "TaskRun") {
		taskStatus = string(v1beta1.TaskRunReasonFailed)
	}
	return &RunningTask{
		Task:       runningTask.Name,
		TaskStatus: taskStatus,
	}
}

func (c *controller) GetContainerLog(ctx context.Context, clusterID uint, podName, containerName string,
	tailLines int) (
	<-chan string, error) {
	const op = "cluster controller: get cluster container log"
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	param := cd.GetContainerLogParams{
		Namespace:   fmt.Sprintf("%v-%v", er.EnvironmentName, application.GroupID),
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Pod:         podName,
		Container:   containerName,
		TailLines:   tailLines,
	}
	return c.cd.GetContainerLog(ctx, &param)
}
