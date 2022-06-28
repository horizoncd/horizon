package cluster

import (
	"context"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/cluster/tekton"
	perror "g.hz.netease.com/horizon/pkg/errors"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
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
	defer wlog.Start(ctx, op).StopPrint()

	resp := &GetClusterStatusResponse{}

	// get latest pipelinerun
	latestPipelinerun, err := c.getLatestPipelinerunByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if latestPipelinerun != nil {
		resp.LatestPipelinerun = &LatestPipelinerun{
			ID:     latestPipelinerun.ID,
			Action: latestPipelinerun.Action,
		}
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	if latestPipelinerun == nil || latestPipelinerun.Action != prmodels.ActionBuildDeploy {
		// if latest builddeploy pr is not exists, runningTask is noneRunningTask
		resp.RunningTask = &RunningTask{
			Task: _taskNone,
		}
	} else {
		latestPipelineRunObject, err := c.getLatestPipelineRunObject(ctx, cluster, latestPipelinerun)
		if err != nil {
			return nil, err
		}

		resp.RunningTask = c.getRunningTask(ctx, latestPipelineRunObject)
		resp.RunningTask.PipelinerunID = latestPipelinerun.ID
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	envValue := &gitrepo.EnvValue{}
	po := &gitrepo.PipelineOutput{}
	if !isClusterStatusUnstable(cluster.Status) {
		envValue, err = c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
		po, err = c.clusterGitRepo.GetPipelineOutput(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil && perror.Cause(err) != herrors.ErrPipelineOutputEmpty {
			return nil, err
		}
	}

	clusterState, err := c.cd.GetClusterState(ctx, &cd.GetClusterStateParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		Namespace:    envValue.Namespace,
		RegionEntity: regionEntity,
	})
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			if cluster.Status != "" {
				resp.ClusterStatus = map[string]string{
					"status": cluster.Status,
				}
			} else {
				resp.ClusterStatus = map[string]string{
					"status": _notFound,
				}
			}
		} else {
			return nil, err
		}
	} else {
		if cluster.Status != "" {
			clusterState.Status = health.HealthStatusCode(cluster.Status)
		}

		// there is a possibility that healthy cluster is not reconciled by operator yet
		if clusterState.Status == health.HealthStatusHealthy {
			if po != nil && po.Image != nil && !isClusterActuallyHealthy(clusterState, *po.Image) {
				clusterState.Status = health.HealthStatusProgressing
			}
		}
		resp.ClusterStatus = clusterState

		if cluster.Status == "" && resp.RunningTask.Task == _taskNone && clusterState.Status != "" {
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

// isClusterActuallyHealthy judge if the cluster is healthy by checking image
func isClusterActuallyHealthy(clusterState *cd.ClusterState, image string) bool {
	checkImage := func(clusterVersion *cd.ClusterVersion) bool {
		if len(clusterVersion.Pods) == 0 || image == "" {
			return true
		}
		for _, pod := range clusterVersion.Pods {
			for _, container := range pod.Spec.InitContainers {
				if container.Image == image {
					return true
				}
			}
			for _, container := range pod.Spec.Containers {
				if container.Image == image {
					return true
				}
			}
		}
		return false
	}

	podTemplateHash := clusterState.PodTemplateHash
	clusterVersion, ok := clusterState.Versions[podTemplateHash]
	if !ok {
		return false
	}

	return checkImage(clusterVersion)
}

func (c *controller) getLatestPipelinerunByClusterID(ctx context.Context,
	clusterID uint) (*prmodels.Pipelinerun, error) {
	_, pipelineruns, err := c.pipelinerunMgr.GetByClusterID(ctx, clusterID, false, q.Query{
		PageNumber: 1,
		PageSize:   1,
	})
	if err != nil {
		return nil, err
	}
	if len(pipelineruns) == 1 {
		return pipelineruns[0], nil
	}
	return nil, nil
}

func (c *controller) getLatestPipelineRunObject(ctx context.Context, cluster *clustermodels.Cluster,
	pipelinerun *prmodels.Pipelinerun) (*v1beta1.PipelineRun, error) {
	var latestPr *v1beta1.PipelineRun
	getPipelineRunFromCollector := func() (*v1beta1.PipelineRun, error) {
		tektonCollector, err := c.tektonFty.GetTektonCollector(cluster.EnvironmentName)
		if err != nil {
			return nil, err
		}
		obj, err := tektonCollector.GetPipelineRunObject(ctx, pipelinerun.PrObject)
		if err != nil {
			return nil, err
		}
		return obj.PipelineRun, nil
	}
	var (
		err          error
		tektonClient tekton.Interface
	)
	if pipelinerun.PrObject == "" {
		tektonClient, err = c.tektonFty.GetTekton(cluster.EnvironmentName)
		if err != nil {
			return nil, err
		}
		latestPr, err = tektonClient.GetPipelineRunByID(ctx, cluster.Name, cluster.ID, pipelinerun.ID)
		if err != nil {
			// get pipelineRun from k8s error, it can be deleted and uploaded to collector,
			// so try to get pipelineRun from collector again
			latestPr, err = getPipelineRunFromCollector()
			if err != nil {
				return nil, err
			}
		}
	} else {
		latestPr, err = getPipelineRunFromCollector()
		if err != nil {
			return nil, err
		}
	}
	return latestPr, nil
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
	if prs.Status == string(v1beta1.PipelineRunReasonTimedOut) {
		taskStatus = string(v1beta1.TaskRunReasonFailed)
	}
	return &RunningTask{
		Task:       runningTask.Name,
		TaskStatus: taskStatus,
	}
}

func (c *controller) GetContainerLog(ctx context.Context, clusterID uint, podName, containerName string,
	tailLines int) (<-chan string, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	param := cd.GetContainerLogParams{
		Namespace:   envValue.Namespace,
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Pod:         podName,
		Container:   containerName,
		TailLines:   tailLines,
	}
	return c.cd.GetContainerLog(ctx, &param)
}

func (c *controller) GetPodEvents(ctx context.Context, clusterID uint, podName string) (interface{}, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	param := cd.GetPodEventsParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		Pod:          podName,
		RegionEntity: regionEntity,
	}
	return c.cd.GetPodEvents(ctx, &param)
}

func (c *controller) GetContainers(ctx context.Context, clusterID uint, podName string) (interface{}, error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	param := cd.GetPodContainersParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		Pod:          podName,
		RegionEntity: regionEntity,
	}
	return c.cd.GetPodContainers(ctx, &param)
}
