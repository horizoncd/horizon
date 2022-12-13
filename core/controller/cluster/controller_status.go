package cluster

import (
	"context"
	"strings"
	"time"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	perror "github.com/horizoncd/horizon/pkg/errors"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"

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

func (c *controller) GetClusterStatusV2(ctx context.Context, clusterID uint) (*StatusResponseV2, error) {
	const op = "cluster controller: get cluster status v2"
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	resp := &StatusResponseV2{}
	// status in db has higher priority
	if cluster.Status != common.ClusterStatusEmpty {
		resp.Status = cluster.Status
	}

	params := &cd.GetClusterStateV2Params{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
	}

	cdStatus, err := c.cd.GetClusterStateV2(ctx, params)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
		// for not deployed -- free or not published
		if resp.Status == "" {
			if cluster.Status != "" {
				resp.Status = cluster.Status
			} else {
				resp.Status = _notFound
			}
		}
	} else {
		// resp has not been set
		if resp.Status == "" {
			resp.Status = string(cdStatus.Status)
		}
	}

	return resp, nil
}

// Deprecated
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

	if latestPipelinerun == nil ||
		latestPipelinerun.Action != prmodels.ActionBuildDeploy {
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

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	clusterState, err := c.cd.GetClusterState(ctx, &cd.GetClusterStateParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
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
		if clusterState.Status == health.HealthStatusHealthy &&
			latestPipelinerun != nil &&
			latestPipelinerun.Status != string(prmodels.StatusFailed) &&
			latestPipelinerun.Status != string(prmodels.StatusCancelled) {
			var (
				image       string
				restartTime time.Time
			)

			// get image and restart time by lastest pipelinerun
			if latestPipelinerun.Action == prmodels.ActionBuildDeploy ||
				latestPipelinerun.Action == prmodels.ActionRollback {
				image = latestPipelinerun.ImageURL
			} else if latestPipelinerun.Action == prmodels.ActionRestart {
				restartTime = latestPipelinerun.CreatedAt
			}

			replicas := 0
			if clusterState.DesiredReplicas != nil {
				replicas = *clusterState.DesiredReplicas
			}

			if !isClusterActuallyHealthy(ctx, clusterState, image,
				restartTime, replicas) {
				clusterState.Status = health.HealthStatusProgressing
			}
		}
		resp.ClusterStatus = clusterState

		if cluster.Status == "" && resp.RunningTask.Task == _taskNone && clusterState.Status != "" {
			// When the task is none, judge type of the pipelinerun (restart, rollback, etc.)
			if clusterState.Status == health.HealthStatusDegraded {
				resp.RunningTask.Task = _taskDeploy
				resp.RunningTask.TaskStatus = string(v1beta1.TaskRunReasonFailed)
			} else if clusterState.Status != health.HealthStatusHealthy {
				// Statuses (Progressing, Suspended) other than Degraded and Healthy are considered to be Running in the release
				resp.RunningTask.Task = _taskDeploy
				resp.RunningTask.TaskStatus = string(v1beta1.TaskRunReasonRunning)
			}
		}
	}
	if cluster.ExpireSeconds > 0 && latestPipelinerun != nil && cluster.Status == "" {
		resp.TTLSeconds = willExpireIn(cluster.ExpireSeconds, cluster.UpdatedAt, latestPipelinerun.UpdatedAt)
	}
	return resp, nil
}

// isClusterActuallyHealthy judge if the cluster is healthy by checking image
func isClusterActuallyHealthy(ctx context.Context, clusterState *cd.ClusterState, image string,
	restartTime time.Time, replicas int) bool {
	checkReplicas := func(clusterVersion *cd.ClusterVersion) bool {
		if replicas == 0 || len(clusterVersion.Pods) == 0 || image == "" {
			return true
		}

		updatedReplicas := 0
		for _, pod := range clusterVersion.Pods {
			containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
			for _, container := range containers {
				if container.Image == image {
					updatedReplicas++
					break
				}
			}
		}

		return updatedReplicas >= replicas
	}

	checkRestart := func(clusterVersion *cd.ClusterVersion) bool {
		if len(clusterVersion.Pods) == 0 || restartTime.IsZero() {
			return true
		}
		checkResult := true
		for podName, pod := range clusterVersion.Pods {
			v, ok := pod.Metadata.Annotations[common.ClusterRestartTimeKey]
			if ok {
				createTime, err := time.Parse("2006-01-02 15:04:05", v)
				if err != nil {
					log.Warningf(ctx, "failed to parse create time of %s: %s, err: %s",
						podName, v, err.Error())
					continue
				}
				if createTime.Before(restartTime) {
					return false
				}
			}
		}
		return checkResult
	}

	podTemplateHash := clusterState.PodTemplateHash
	clusterVersion, ok := clusterState.Versions[podTemplateHash]
	if !ok {
		return false
	}

	return checkReplicas(clusterVersion) && checkRestart(clusterVersion)
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
		latestPr, err = tektonClient.GetPipelineRunByID(ctx, pipelinerun.CIEventID)
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				return nil, nil
			}
			return nil, err
		}
	} else {
		latestPr, err = getPipelineRunFromCollector()
		if err != nil {
			return nil, err
		}
	}
	return latestPr, nil
}

// getRunningTask Get the latest currently executing Task of the pipeline Run.
// If the last executed pipelineRun is successful,
// return noneRunningTask
func (c *controller) getRunningTask(ctx context.Context, pr *v1beta1.PipelineRun) *RunningTask {
	noneRunningTask := &RunningTask{
		Task: _taskNone,
	}
	if pr == nil {
		// if tekton pipelinerun is not created yet, return pending build task
		return &RunningTask{
			Task:       _taskBuild,
			TaskStatus: _taskStatusPending,
		}
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
		// If the pipeline is in a completed state, none Running Task will be returned.
		// Indicates that there is no currently executing Task
		return noneRunningTask
	}

	runningTask := prs.RunningTask
	if runningTask == nil {
		// If the pipeline has not started execution, the return build is Pending status
		return &RunningTask{
			Task:       _taskBuild,
			TaskStatus: _taskStatusPending,
		}
	}
	taskStatus := strings.TrimPrefix(runningTask.Status, "TaskRun")
	// If it is Timeout, it is considered Failed
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

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
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

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
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

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	param := cd.GetPodParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		Pod:          podName,
		RegionEntity: regionEntity,
	}
	return c.cd.GetPodContainers(ctx, &param)
}

func (c *controller) GetClusterPod(ctx context.Context, clusterID uint, podName string) (
	*GetClusterPodResponse, error) {
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

	param := cd.GetPodParams{
		Namespace:    envValue.Namespace,
		Cluster:      cluster.Name,
		Pod:          podName,
		RegionEntity: regionEntity,
	}
	pod, err := c.cd.GetPod(ctx, &param)
	if err != nil {
		return nil, err
	}
	return &GetClusterPodResponse{Pod: *pod}, nil
}

func (c *controller) GetResourceTree(ctx context.Context, clusterID uint) (resp *GetResourceTreeResponse, err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	resourceTree, err := c.cd.GetResourceTree(ctx, &cd.GetResourceTreeParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
	})
	if err != nil {
		return
	}

	resp = &GetResourceTreeResponse{}
	resp.Nodes = make(map[string]*ResourceNode, len(resourceTree))
	for i := range resourceTree {
		n := ResourceNode{
			ResourceNode: resourceTree[i].ResourceNode,
			PodDetail:    resourceTree[i].PodDetail,
		}
		resp.Nodes[n.UID] = &n
	}
	return
}

func (c *controller) GetStep(ctx context.Context, clusterID uint) (resp *GetStepResponse, err error) {
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	steps, err := c.cd.GetStep(ctx, &cd.GetStepParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
	})
	if err != nil {
		return
	}

	if steps != nil {
		resp = &GetStepResponse{
			Total:        steps.Total,
			Index:        steps.Index,
			Replicas:     steps.Replicas,
			ManualPaused: steps.ManualPaused,
		}
	} else {
		resp = &GetStepResponse{
			Total:        0,
			Index:        0,
			Replicas:     []int{},
			ManualPaused: false,
		}
	}
	return
}

func willExpireIn(ttl uint, tms ...time.Time) *uint {
	var (
		latestTime time.Time
		res        uint
	)
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}

	expireAt := latestTime.Add(time.Duration(ttl) * time.Second)
	if expireAt.Before(time.Now()) {
		res = uint(0)
		return &res
	}
	res = uint(time.Until(expireAt).Seconds())
	return &res
}
