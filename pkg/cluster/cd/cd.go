/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cd

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"

	herrors "g.hz.netease.com/horizon/core/errors"

	"g.hz.netease.com/horizon/pkg/cluster/cd/argocd"
	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	argocdconf "g.hz.netease.com/horizon/pkg/config/argocd"
	perror "g.hz.netease.com/horizon/pkg/errors"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	v1alpha12 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	kube2 "github.com/argoproj/gitops-engine/pkg/utils/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/describe"
	kubectlresource "k8s.io/kubectl/pkg/util/resource"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"
)

const (
	// PodLifeCycleSchedule specifies whether pod has been scheduled
	PodLifeCycleSchedule = "PodSchedule"
	// PodLifeCycleInitialize specifies whether all init containers have finished
	PodLifeCycleInitialize = "PodInitialize"
	// PodLifeCycleContainerStartup specifies whether the container has passed its startup probe
	PodLifeCycleContainerStartup = "ContainerStartup"
	// PodLifeCycleContainerOnline specified whether the container has passed its postStart hook
	PodLifeCycleContainerOnline = "ContainerOnline"
	// PodLifeCycleHealthCheck specifies whether the container has passed its readiness probe
	PodLifeCycleHealthCheck = "HealthCheck"
	// PodLifeCycleContainerPreStop specifies whether the container is executing preStop hook
	PodLifeCycleContainerPreStop = "PreStop"

	LifeCycleStatusSuccess  = "Success"
	LifeCycleStatusWaiting  = "Waiting"
	LifeCycleStatusRunning  = "Running"
	LifeCycleStatusAbnormal = "Abnormal"

	PodErrCrashLoopBackOff = "CrashLoopBackOff"
)

type GetClusterStateParams struct {
	Environment  string
	Cluster      string
	RegionEntity *regionmodels.RegionEntity
}

type CreateClusterParams struct {
	Environment  string
	Cluster      string
	GitRepoURL   string
	ValueFiles   []string
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
}

type DeployClusterParams struct {
	Environment string
	Cluster     string
	Revision    string
}

type GetPodEventsParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Pod          string
}

type GetPodParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Pod          string
}

type DeletePodsParams struct {
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
	Pods         []string
}

type DeleteClusterParams struct {
	Environment string
	Cluster     string
}

type ClusterNextParams struct {
	Environment string
	Cluster     string
}

type ClusterPromoteParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type ClusterPauseParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type ClusterResumeParams struct {
	RegionEntity *regionmodels.RegionEntity
	Cluster      string
	Namespace    string
	Environment  string
}

type GetContainerLogParams struct {
	Namespace   string
	Cluster     string
	Pod         string
	Container   string
	Environment string
	TailLines   int
}

type ExecParams struct {
	Environment  string
	Cluster      string
	RegionEntity *regionmodels.RegionEntity
	Namespace    string
	PodList      []string
}

type ExecResp struct {
	key    string
	Result bool
	Stdout string
	Stderr string
	Error  error
}

type OperationResult struct {
	Result bool
	Error  error
}

type ExecFunc func(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)

type CD interface {
	CreateCluster(ctx context.Context, params *CreateClusterParams) error
	DeployCluster(ctx context.Context, params *DeployClusterParams) error
	DeleteCluster(ctx context.Context, params *DeleteClusterParams) error
	Next(ctx context.Context, params *ClusterNextParams) error
	Promote(ctx context.Context, params *ClusterPromoteParams) error
	Pause(ctx context.Context, params *ClusterPauseParams) error
	Resume(ctx context.Context, params *ClusterResumeParams) error
	// GetClusterState get cluster state in cd system
	GetClusterState(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
	GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error)
	GetPod(ctx context.Context, params *GetPodParams) (*corev1.Pod, error)
	GetPodContainers(ctx context.Context, params *GetPodParams) ([]ContainerDetail, error)
	GetPodEvents(ctx context.Context, params *GetPodEventsParams) ([]Event, error)
	Online(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	Offline(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	DeletePods(ctx context.Context, params *DeletePodsParams) (map[string]OperationResult, error)
}

type cd struct {
	kubeClientFty kubeclient.Factory
	factory       argocd.Factory
}

func NewCD(argoCDMapper argocdconf.Mapper) CD {
	return &cd{
		kubeClientFty: kubeclient.Fty,
		factory:       argocd.NewFactory(argoCDMapper),
	}
}

func (c *cd) CreateCluster(ctx context.Context, params *CreateClusterParams) (err error) {
	const op = "cd: create cluster"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return err
	}

	// if argo application exists, return, else create it
	_, err = argo.GetApplication(ctx, params.Cluster)
	if err == nil {
		return nil
	}
	if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
		return err
	}
	var argoApplication = argocd.AssembleArgoApplication(params.Cluster, params.Namespace,
		params.GitRepoURL, params.RegionEntity.Server, params.ValueFiles)

	manifest, err := json.Marshal(argoApplication)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	if err := argo.CreateApplication(ctx, manifest); err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return nil
}

func (c *cd) DeployCluster(ctx context.Context, params *DeployClusterParams) (err error) {
	const op = "cd: deploy cluster"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return argo.DeployApplication(ctx, params.Cluster, params.Revision)
}

func (c *cd) DeleteCluster(ctx context.Context, params *DeleteClusterParams) (err error) {
	const op = "cd: delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return err
	}

	// 1. get application first
	applicationCR, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil
		}
		return
	}

	// 2. delete application
	if err := argo.DeleteApplication(ctx, params.Cluster); err != nil {
		return err
	}

	// 3. wait for application to delete completely
	return argo.WaitApplication(ctx, params.Cluster, string(applicationCR.UID), http.StatusNotFound)
}

func (c *cd) Next(ctx context.Context, params *ClusterNextParams) (err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return err
	}

	return argo.ResumeRollout(ctx, params.Cluster)
}

var rolloutResource = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "rollouts",
}

// Promote a paused rollout
func (c *cd) Promote(ctx context.Context, params *ClusterPromoteParams) (err error) {
	// 1. get argo rollout
	rollout, err := c.getRollout(ctx, params.Environment, params.Cluster)
	if err != nil {
		return err
	}
	if rollout == nil {
		return nil
	}

	// 2. patch rollout
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return err
	}
	patchBody := []byte(getSkipAllStepsPatchStr(len(rollout.Spec.Strategy.Canary.Steps)))
	_, err = kubeClient.Dynamic.Resource(rolloutResource).
		Namespace(params.Namespace).
		Patch(ctx, params.Cluster, types.MergePatchType, patchBody, metav1.PatchOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

// Pause a rollout
func (c *cd) Pause(ctx context.Context, params *ClusterPauseParams) (err error) {
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return perror.WithMessagef(err, "failed to get argocd application resource for cluster %s",
			params.Cluster)
	}
	patchBody := []byte(getPausePatchStr())
	_, err = kubeClient.Dynamic.Resource(rolloutResource).
		Namespace(params.Namespace).
		Patch(ctx, params.Cluster, types.MergePatchType, patchBody, metav1.PatchOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

// Resume a paused rollout
func (c *cd) Resume(ctx context.Context, params *ClusterResumeParams) (err error) {
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return perror.WithMessagef(err, "failed to get argocd application resource for cluster %s",
			params.Cluster)
	}
	patchBody := []byte(getResumePatchStr())
	_, err = kubeClient.Dynamic.Resource(rolloutResource).
		Namespace(params.Namespace).
		Patch(ctx, params.Cluster, types.MergePatchType, patchBody, metav1.PatchOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

func (c *cd) getRollout(ctx context.Context, environment, clusterName string) (*v1alpha1.Rollout, error) {
	var rollout *v1alpha1.Rollout
	argo, err := c.factory.GetArgoCD(environment)
	if err != nil {
		return nil, err
	}
	argoApp, err := argo.GetApplication(ctx, clusterName)
	if err != nil {
		return nil, perror.WithMessagef(err, "failed to get argocd application: %s", clusterName)
	}
	if err := argo.GetApplicationResource(ctx, clusterName, argocd.ResourceParams{
		Group:        "argoproj.io",
		Version:      "v1alpha1",
		Kind:         "Rollout",
		Namespace:    argoApp.Spec.Destination.Namespace,
		ResourceName: clusterName,
	}, &rollout); err != nil {
		if perror.Cause(err) == argocd.ErrResourceNotFound {
			return nil, nil
		}
		return nil, perror.WithMessagef(err, "failed to get argocd application resource for cluster %s",
			clusterName)
	}

	return rollout, nil
}

// GetClusterState TODO(gjq) restructure
func (c *cd) GetClusterState(ctx context.Context,
	params *GetClusterStateParams) (clusterState *ClusterState, err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	clusterState = &ClusterState{Versions: map[string]*ClusterVersion{}}

	// get application status
	argoApp, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}
	namespace := argoApp.Spec.Destination.Namespace

	// namespace = argoApp.Spec.Destination.Namespace
	clusterState.Status = argoApp.Status.Health.Status
	if clusterState.Status == "" {
		return nil, herrors.NewErrNotFound(herrors.ClusterStateInArgo, "clusterState.State == \"\"")
	}
	if clusterState.Status == health.HealthStatusUnknown {
		clusterState.Status = health.HealthStatusDegraded
	} else if clusterState.Status == health.HealthStatusMissing {
		clusterState.Status = health.HealthStatusProgressing
	}

	var rollout *v1alpha1.Rollout
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterLabelKey, params.Cluster))
	if err := argo.GetApplicationResource(ctx, params.Cluster, argocd.ResourceParams{
		Group:        "argoproj.io",
		Version:      "v1alpha1",
		Kind:         "Rollout",
		Namespace:    argoApp.Spec.Destination.Namespace,
		ResourceName: params.Cluster,
	}, &rollout); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			// get pods by resourceTree
			var (
				clusterPodMap = map[string]*ClusterPod{}
				podMap        = map[string]corev1.Pod{}
			)
			resourceTree, err := argo.GetApplicationTree(ctx, params.Cluster)
			if err != nil {
				return nil, err
			}
			// application with deployment may be serverless
			if !resourceTreeContains(resourceTree, kube2.DeploymentKind) {
				allPods, err := kube.GetPods(ctx, kubeClient.Basic, namespace, labelSelector.String())
				if err != nil {
					return nil, err
				}
				for _, pod := range allPods {
					podMap[pod.Name] = pod
				}
				for _, node := range resourceTree.Nodes {
					if node.Kind == kube2.PodKind {
						if _, ok := podMap[node.Name]; !ok {
							return nil, herrors.NewErrNotFound(herrors.PodsInK8S, fmt.Sprintf("pod %s does not exist", node.Name))
						}
						clusterPodMap[node.Name] = podToClusterPod(podMap[node.Name])
					}
				}
				clusterState.PodTemplateHash = "default"
				clusterState.PodTemplateHashKey = "default"
				clusterState.Replicas = len(clusterPodMap)
				clusterState.Versions["default"] = &ClusterVersion{
					Replicas: len(clusterPodMap),
					Pods:     clusterPodMap,
				}
				clusterState.Step = &Step{
					Index:    0,
					Total:    1,
					Replicas: []int{1},
				}
				return clusterState, nil
			}
		} else {
			return nil, perror.WithMessagef(err,
				"failed to get rollout for cluster %s", params.Cluster)
		}
	}
	clusterState.Step = getStep(rollout)
	if rollout != nil {
		desiredReplicas := 1
		if rollout.Spec.Replicas != nil {
			desiredReplicas = int(*rollout.Spec.Replicas)
		}
		clusterState.DesiredReplicas = &desiredReplicas
		clusterState.ManualPaused = rollout.Spec.Paused
	}

	var latestReplicaSet *appsv1.ReplicaSet
	rss, err := kube.GetReplicaSets(ctx, kubeClient.Basic, namespace, labelSelector.String())
	if err != nil {
		return nil, err
	} else if len(rss) == 0 {
		return nil, herrors.NewErrNotFound(herrors.ReplicasSetInK8S, "ReplicaSet instance not found")
	}

	for i := range rss {
		rs := &rss[i]

		if latestReplicaSet == nil || CompareRevision(ctx, rs, latestReplicaSet) {
			latestReplicaSet = rs
		}
		_, hash := getPodTemplateHash(rs)
		clusterState.Versions[hash] = &ClusterVersion{
			Pods:     map[string]*ClusterPod{},
			Revision: getRevision(rs),
		}
	}

	// set revision, podTemplateHash
	clusterState.PodTemplateHashKey, clusterState.PodTemplateHash = getPodTemplateHash(latestReplicaSet)
	clusterState.Revision = getRevision(latestReplicaSet)

	if clusterState.PodTemplateHash == "" {
		return nil, herrors.NewErrNotFound(herrors.ClusterStateInArgo, "clusterState.PodTemplateHash == ''")
	}

	// TODO(gjq): 通用化，POD的展示是直接按照resourceVersion 来获取Pod
	// 从目前的配置来看，该 if 分支表示负载类型是 serverless 应用
	if clusterState.PodTemplateHashKey == DeploymentPodTemplateHash {
		labelSelector := fields.ParseSelectorOrDie(
			fmt.Sprintf("%v=%v", common.ClusterLabelKey, params.Cluster))
		// serverless 应用会有多个 Deployment 对象
		deploymentList, err := kube.GetDeploymentList(ctx, kubeClient.Basic, namespace, labelSelector.String())
		if err != nil {
			return nil, err
		}
		var latestDeployment *appsv1.Deployment
		for i := range deploymentList {
			if latestDeployment == nil ||
				deploymentList[i].CreationTimestamp.After(latestDeployment.CreationTimestamp.Time) {
				latestDeployment = &deploymentList[i]
			}
		}
		// Borrowed at kubernetes/kubectl/rollout_status.go
		if latestDeployment != nil {
			if latestDeployment.Generation <= latestDeployment.Status.ObservedGeneration {
				cond := getDeploymentCondition(latestDeployment.Status, appsv1.DeploymentProgressing)
				if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
					// 默认情况下，如果 Deployment 在十分钟之内，未能完成滚动更新，
					// 则 Deployment 的健康状态是 HealthStatusDegraded.
					clusterState.Status = health.HealthStatusDegraded
				}
			} else {
				// 如果 Deployment 有更新，而 Deployment Controller 尚未处理，
				// 则 Deployment 的健康状态是 HealthStatusProgressing
				clusterState.Status = health.HealthStatusProgressing
			}
		}
	}

	if err := c.paddingPodAndEventInfo(ctx, params.Cluster, namespace,
		kubeClient.Basic, clusterState); err != nil {
		return nil, err
	}

	return clusterState, nil
}

func (c *cd) GetPodContainers(ctx context.Context,
	params *GetPodParams) (containers []ContainerDetail, err error) {
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	pod, err := kube.GetPod(ctx, kubeClient.Basic, params.Namespace, params.Pod)
	if err != nil {
		return nil, err
	}

	return extractContainerDetail(pod), nil
}

func (c *cd) GetPod(ctx context.Context,
	params *GetPodParams) (pod *corev1.Pod, err error) {
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	pod, err = kube.GetPod(ctx, kubeClient.Basic, params.Namespace, params.Pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (c *cd) DeletePods(ctx context.Context,
	params *DeletePodsParams) (map[string]OperationResult, error) {
	result := map[string]OperationResult{}
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return result, err
	}

	for _, pod := range params.Pods {
		err = kube.DeletePods(ctx, kubeClient.Basic, params.Namespace, pod)
		if err != nil {
			result[pod] = OperationResult{
				Result: false,
				Error:  err,
			}
			continue
		}
		result[pod] = OperationResult{
			Result: true,
		}
	}

	return result, nil
}

func (c *cd) GetPodEvents(ctx context.Context,
	params *GetPodEventsParams) (events []Event, err error) {
	const op = "cd: get cluster pod events"
	defer wlog.Start(ctx, op).StopPrint()

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterLabelKey, params.Cluster))
	pods, err := kube.GetPods(ctx, kubeClient.Basic, params.Namespace, labelSelector.String())
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if pod.Name == params.Pod {
			k8sEvents, err := kube.GetPodEvents(ctx, kubeClient.Basic, params.Namespace, params.Pod)
			if err != nil {
				return nil, err
			}

			for _, event := range k8sEvents {
				eventTimeStamp := metav1.Time{Time: event.EventTime.Time}
				if eventTimeStamp.IsZero() {
					eventTimeStamp = event.FirstTimestamp
				}
				events = append(events, Event{
					Type:           event.Type,
					Reason:         event.Reason,
					Message:        event.Message,
					Count:          event.Count,
					EventTimestamp: eventTimeStamp,
				})
			}
			return events, nil
		}
	}

	return nil, herrors.NewErrNotFound(herrors.PodsInK8S, "pod does not exist")
}

// extractContainerInfo extract container detail
func extractContainerDetail(pod *corev1.Pod) []ContainerDetail {
	containers := make([]ContainerDetail, 0)
	for _, container := range pod.Spec.Containers {
		vars := extractEnv(pod, container)
		volumeMounts := extractContainerMounts(container, pod)

		containers = append(containers, ContainerDetail{
			Name:            container.Name,
			Image:           container.Image,
			Env:             vars,
			Commands:        container.Command,
			Args:            container.Args,
			VolumeMounts:    volumeMounts,
			SecurityContext: container.SecurityContext,
			Status:          extractContainerStatus(pod, &container),
			LivenessProbe:   container.LivenessProbe,
			ReadinessProbe:  container.ReadinessProbe,
			StartupProbe:    container.StartupProbe,
		})
	}
	return containers
}

// extractContainerMounts extract container status from pod.status.containerStatus
func extractContainerStatus(pod *corev1.Pod, container *corev1.Container) *corev1.ContainerStatus {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == container.Name {
			return &status
		}
	}
	return nil
}

// extractContainerMounts extract container mounts
// the same to https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/detail.go#L226
func extractContainerMounts(container corev1.Container, pod *corev1.Pod) []VolumeMount {
	volumeMounts := make([]VolumeMount, 0)
	for _, volumeMount := range container.VolumeMounts {
		volumeMounts = append(volumeMounts, VolumeMount{
			Name:      volumeMount.Name,
			ReadOnly:  volumeMount.ReadOnly,
			MountPath: volumeMount.MountPath,
			SubPath:   volumeMount.SubPath,
			Volume:    getVolume(pod.Spec.Volumes, volumeMount.Name),
		})
	}
	return volumeMounts
}

// getVolume get volume by name
// the same to https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/detail.go#L216
func getVolume(volumes []corev1.Volume, volumeName string) corev1.Volume {
	for _, volume := range volumes {
		if volume.Name == volumeName {
			return volume
		}
	}
	return corev1.Volume{}
}

// extractEnv extract env by resolving references
// the same to https://github.com/kubernetes/kubectl/blob/master/pkg/describe/describe.go#L1853
// todo: maybe we should follow dashboard to resolve config/secret references
// https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/detail.go#L303
func extractEnv(pod *corev1.Pod, container corev1.Container) []corev1.EnvVar {
	var env []corev1.EnvVar
	for _, e := range container.Env {
		switch {
		case e.ValueFrom == nil:
			env = append(env, e)
		case e.ValueFrom.FieldRef != nil:
			var valueFrom string
			valueFrom = describe.EnvValueRetriever(pod)(e)
			env = append(env, corev1.EnvVar{
				Name:  e.Name,
				Value: valueFrom,
			})
		case e.ValueFrom.ResourceFieldRef != nil:
			valueFrom, err := kubectlresource.ExtractContainerResourceValue(e.ValueFrom.ResourceFieldRef, &container)
			if err != nil {
				valueFrom = ""
			}
			resource := e.ValueFrom.ResourceFieldRef.Resource
			if valueFrom == "0" && (resource == "limits.cpu" || resource == "limits.memory") {
				valueFrom = "node allocatable"
			}
			env = append(env, corev1.EnvVar{
				Name:  e.Name,
				Value: valueFrom,
			})
		case e.ValueFrom.SecretKeyRef != nil:
			optional := e.ValueFrom.SecretKeyRef.Optional != nil && *e.ValueFrom.SecretKeyRef.Optional
			env = append(env, corev1.EnvVar{
				Name: e.Name,
				Value: fmt.Sprintf("<set to the key '%s' in secret '%s'>\tOptional: %t\n",
					e.ValueFrom.SecretKeyRef.Key, e.ValueFrom.SecretKeyRef.Name, optional),
			})
		case e.ValueFrom.ConfigMapKeyRef != nil:
			optional := e.ValueFrom.ConfigMapKeyRef.Optional != nil && *e.ValueFrom.ConfigMapKeyRef.Optional
			env = append(env, corev1.EnvVar{
				Name: e.Name,
				Value: fmt.Sprintf("<set to the key '%s' of config map '%s'>\tOptional: %t\n",
					e.ValueFrom.ConfigMapKeyRef.Key, e.ValueFrom.ConfigMapKeyRef.Name, optional),
			})
		}
	}
	return env
}

func (c *cd) paddingPodAndEventInfo(ctx context.Context, cluster, namespace string,
	kubeClientSet kubernetes.Interface, clusterState *ClusterState) error {
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterLabelKey, cluster))

	var pods []corev1.Pod
	var events map[string][]*corev1.Event
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pods, err1 = kube.GetPods(ctx, kubeClientSet, namespace, labelSelector.String())
	}()
	go func() {
		defer wg.Done()
		events, err2 = kube.GetEvents(ctx, kubeClientSet, namespace)
	}()
	wg.Wait()

	for _, e := range []error{err1, err2} {
		if e != nil {
			return e
		}
	}

	for i := range pods {
		pod := &pods[i]
		podEvents := events[fmt.Sprintf("%v-%v-%v", pod.Name, pod.UID, pod.Namespace)]
		if err := parsePod(ctx, clusterState, pod, podEvents); err != nil {
			// pod 可能已经被删除
			log.Info(ctx, err)
			continue
		} else {
			clusterState.Replicas++
		}
	}
	return nil
}

func resourceTreeContains(resourceTree *v1alpha12.ApplicationTree, resourceKind string) bool {
	for _, node := range resourceTree.Nodes {
		if node.Kind == resourceKind {
			return true
		}
	}
	return false
}

func podToClusterPod(pod corev1.Pod) (clusterPod *ClusterPod) {
	clusterPod = &ClusterPod{
		Metadata: PodMetadata{
			CreationTimestamp: pod.CreationTimestamp,
			Namespace:         pod.Namespace,
			Annotations:       pod.Annotations,
		},
		Spec: PodSpec{
			NodeName:       pod.Spec.NodeName,
			InitContainers: nil,
			Containers:     nil,
		},
		Status: PodStatus{
			HostIP: pod.Status.HostIP,
			PodIP:  pod.Status.PodIP,
			Phase:  string(pod.Status.Phase),
		},
		DeletionTimestamp: pod.DeletionTimestamp,
	}

	var initContainers []*Container
	for i := range pod.Spec.InitContainers {
		c := pod.Spec.InitContainers[i]
		initContainers = append(initContainers, &Container{
			Name:  c.Name,
			Image: c.Image,
		})
	}
	clusterPod.Spec.InitContainers = initContainers

	cs := &containerList{}
	for i := range pod.Spec.Containers {
		cs.containers = append(cs.containers, &pod.Spec.Containers[i])
	}
	sort.Sort(cs)

	var containers []*Container
	for i := range cs.containers {
		c := cs.containers[i]
		containers = append(containers, &Container{
			Name:  c.Name,
			Image: c.Image,
		})
	}
	clusterPod.Spec.Containers = containers

	var containerStatuses []*ContainerStatus
	for i := range pod.Status.ContainerStatuses {
		containerStatus := pod.Status.ContainerStatuses[i]
		c := &ContainerStatus{
			Name:         containerStatus.Name,
			Ready:        containerStatus.Ready,
			RestartCount: containerStatus.RestartCount,
			State:        parseContainerState(containerStatus),
			ImageID:      containerStatus.ImageID,
		}
		containerStatuses = append(containerStatuses, c)
	}
	clusterPod.Status.ContainerStatuses = containerStatuses
	clusterPod.Status.LifeCycle = parsePodLifeCycle(pod)
	return
}

func parsePod(ctx context.Context, clusterInfo *ClusterState,
	pod *corev1.Pod, events []*corev1.Event) (err error) {
	const deploymentPodTemplateHash = "pod-template-hash"
	const rolloutPodTemplateHash = "rollouts-pod-template-hash"

	podTemplateHash := pod.Labels[deploymentPodTemplateHash]
	if podTemplateHash == "" {
		podTemplateHash = pod.Labels[rolloutPodTemplateHash]
	}

	if podTemplateHash == "" {
		log.Errorf(ctx, "pod<%s> has no %v or %v label",
			pod.Name, deploymentPodTemplateHash, rolloutPodTemplateHash)
		return nil
	}

	if clusterInfo.Versions[podTemplateHash] == nil {
		log.Infof(ctx, "pod<%s> has no related ReplicaSet object", pod.Name)
		return nil
	}
	clusterInfo.Versions[podTemplateHash].Replicas++

	clusterPod := &ClusterPod{
		Metadata: PodMetadata{
			CreationTimestamp: pod.CreationTimestamp,
			Namespace:         pod.Namespace,
			Annotations:       pod.Annotations,
		},
		Spec: PodSpec{
			NodeName:       pod.Spec.NodeName,
			InitContainers: nil,
			Containers:     nil,
		},
		Status: PodStatus{
			HostIP: pod.Status.HostIP,
			PodIP:  pod.Status.PodIP,
			Phase:  string(pod.Status.Phase),
		},
		DeletionTimestamp: pod.DeletionTimestamp,
	}

	var initContainers []*Container
	for i := range pod.Spec.InitContainers {
		c := pod.Spec.InitContainers[i]
		initContainers = append(initContainers, &Container{
			Name:  c.Name,
			Image: c.Image,
		})
	}
	clusterPod.Spec.InitContainers = initContainers

	cs := &containerList{name: pod.Labels[common.ClusterLabelKey]}
	for i := range pod.Spec.Containers {
		cs.containers = append(cs.containers, &pod.Spec.Containers[i])
	}
	sort.Sort(cs)

	var containers []*Container
	// containerNameList holds the containerName order
	containerNameList := make([]string, 0)
	for i := range cs.containers {
		c := cs.containers[i]
		containers = append(containers, &Container{
			Name:  c.Name,
			Image: c.Image,
		})
		containerNameList = append(containerNameList, c.Name)
	}
	clusterPod.Spec.Containers = containers

	// containerStatusMap, key is containerName, value is *ContainerStatus
	containerStatusMap := make(map[string]*ContainerStatus)
	for i := range pod.Status.ContainerStatuses {
		containerStatus := pod.Status.ContainerStatuses[i]
		c := &ContainerStatus{
			Name:         containerStatus.Name,
			Ready:        containerStatus.Ready,
			RestartCount: containerStatus.RestartCount,
			State:        parseContainerState(containerStatus),
			ImageID:      containerStatus.ImageID,
		}
		containerStatusMap[containerStatus.Name] = c
	}

	// construct ContainerStatus list, in containerName order
	var containerStatuses []*ContainerStatus
	for _, containerName := range containerNameList {
		if c, ok := containerStatusMap[containerName]; ok {
			containerStatuses = append(containerStatuses, c)
			delete(containerStatusMap, containerName)
		}
	}
	// append the rest ContainerStatus in containerStatusMap if it exists
	for containerName := range containerStatusMap {
		containerStatuses = append(containerStatuses, containerStatusMap[containerName])
	}

	clusterPod.Status.ContainerStatuses = containerStatuses
	clusterPod.Status.LifeCycle = parsePodLifeCycle(*pod)

	for i := range events {
		eventTimeStamp := metav1.Time{Time: events[i].EventTime.Time}
		if eventTimeStamp.IsZero() {
			eventTimeStamp = events[i].FirstTimestamp
		}
		clusterPod.Status.Events = append(clusterPod.Status.Events,
			Event{
				Type:           events[i].Type,
				Reason:         events[i].Reason,
				Message:        events[i].Message,
				Count:          events[i].Count,
				EventTimestamp: eventTimeStamp,
			})
	}
	clusterInfo.Versions[podTemplateHash].Pods[pod.Name] = clusterPod
	return nil
}

type containerList struct {
	name       string
	containers []*corev1.Container
}

func (c *containerList) Len() int { return len(c.containers) }

func (c *containerList) Less(i, j int) bool {
	if c.containers[i].Name == c.name {
		return true
	} else if c.containers[j].Name == c.name {
		return false
	}

	a := c.containers[i].Name
	b := c.containers[j].Name

	return a <= b
}

func (c *containerList) Swap(i, j int) {
	c.containers[i], c.containers[j] = c.containers[j], c.containers[i]
}

// parsePodLifecycle parse pod lifecycle by pod status
func parsePodLifeCycle(pod corev1.Pod) []*LifeCycleItem {
	var lifeCycle []*LifeCycleItem
	// if DeletionTimestamp is set, pod is Terminating
	if pod.DeletionTimestamp != nil {
		lifeCycle = []*LifeCycleItem{
			{
				Type:   PodLifeCycleContainerPreStop,
				Status: LifeCycleStatusRunning,
			},
		}
	} else {
		status := pod.Status
		var (
			conditionMap = map[corev1.PodConditionType]corev1.PodCondition{}
			schedule     = LifeCycleItem{
				Type:   PodLifeCycleSchedule,
				Status: LifeCycleStatusWaiting,
			}
			initialize = LifeCycleItem{
				Type:   PodLifeCycleInitialize,
				Status: LifeCycleStatusWaiting,
			}
			containerStartup = LifeCycleItem{
				Type:   PodLifeCycleContainerStartup,
				Status: LifeCycleStatusWaiting,
			}
			containerOnline = LifeCycleItem{
				Type:   PodLifeCycleContainerOnline,
				Status: LifeCycleStatusWaiting,
			}
			healthCheck = LifeCycleItem{
				Type:   PodLifeCycleHealthCheck,
				Status: LifeCycleStatusWaiting,
			}
		)
		lifeCycle = []*LifeCycleItem{
			&schedule,
			&initialize,
			&containerStartup,
			&containerOnline,
			&healthCheck,
		}
		if len(status.ContainerStatuses) == 0 {
			return lifeCycle
		}

		for _, condition := range status.Conditions {
			conditionMap[condition.Type] = condition
		}
		if condition, ok := conditionMap[corev1.PodScheduled]; ok {
			if condition.Status == corev1.ConditionTrue {
				schedule.Status = LifeCycleStatusSuccess
				schedule.CompleteTime = condition.LastTransitionTime
				initialize.Status = LifeCycleStatusRunning
			} else if condition.Message != "" {
				schedule.Status = LifeCycleStatusAbnormal
				schedule.Message = condition.Message
			}
		} else {
			schedule.Status = LifeCycleStatusWaiting
		}

		if condition, ok := conditionMap[corev1.PodInitialized]; ok {
			if condition.Status == corev1.ConditionTrue {
				initialize.Status = LifeCycleStatusSuccess
				initialize.CompleteTime = condition.LastTransitionTime
				containerStartup.Status = LifeCycleStatusRunning
			}
		} else {
			initialize.Status = LifeCycleStatusWaiting
		}

		if allContainersStarted(status.ContainerStatuses) {
			containerStartup.Status = LifeCycleStatusSuccess
			containerOnline.Status = LifeCycleStatusRunning
		}

		if allContainersRunning(status.ContainerStatuses) {
			containerOnline.Status = LifeCycleStatusSuccess
			healthCheck.Status = LifeCycleStatusRunning
		}

		if allContainersReady(status.ContainerStatuses) {
			healthCheck.Status = LifeCycleStatusSuccess
		}

		// CrashLoopBackOff means rest items are abnormal
		if oneOfContainersCrash(status.ContainerStatuses) {
			for i := 0; i < len(lifeCycle); i++ {
				if lifeCycle[i].Status == LifeCycleStatusRunning {
					lifeCycle[i].Status = LifeCycleStatusAbnormal
				}
			}
		}
	}

	return lifeCycle
}

// allContainersStarted determine if all containers have been started
func allContainersStarted(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Started == nil || !*(containerStatus.Started) {
			return false
		}
	}
	return true
}

// allContainersRunning determine if all containers running
func allContainersRunning(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.State.Running == nil {
			return false
		}
	}
	return true
}

// allContainersReady determine if all containers ready
func allContainersReady(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return true
}

// oneOfContainersCrash determine if one of containers crash
func oneOfContainersCrash(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == PodErrCrashLoopBackOff {
			return true
		}
	}
	return false
}

func parseContainerState(containerStatus corev1.ContainerStatus) ContainerState {
	waiting := "waiting"
	running := "running"
	terminated := "terminated"

	// Only one of its members may be specified.
	state := containerStatus.State

	if state.Running != nil {
		return ContainerState{
			State:     running,
			StartedAt: &state.Running.StartedAt,
		}
	}

	if state.Waiting != nil {
		return ContainerState{
			State:   waiting,
			Reason:  state.Waiting.Reason,
			Message: state.Waiting.Message,
		}
	}

	if state.Terminated != nil {
		return ContainerState{
			State:     terminated,
			Reason:    state.Terminated.Reason,
			Message:   state.Terminated.Message,
			StartedAt: &state.Terminated.StartedAt,
		}
	}

	// If none of them is specified, the default one is ContainerStateWaiting.
	return ContainerState{State: waiting}
}

type (

	// ClusterState 集群状态信息
	ClusterState struct {
		// Status:
		// Processing(正在部署)；Healthy(部署完成)
		// Suspended(已暂停)；Degraded(已降级)
		// Missing(Rollout或Deployment还尚未部署到业务集群)
		// Unknown(集群健康评估失败，无法获悉当前的部署状态)
		Status health.HealthStatusCode `json:"status,omitempty" yaml:"status,omitempty"`

		// Step
		Step *Step `json:"step"`

		// Replicas the actual number of replicas running in k8s
		Replicas int `json:"replicas,omitempty" yaml:"replicas,omitempty"`

		// DesiredReplicas desired replicas
		DesiredReplicas *int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`

		// PodTemplateHash
		PodTemplateHash string `json:"podTemplateHash,omitempty" yaml:"podTemplateHash,omitempty"`

		// PodTemplateHashKey 在 Deployment 或 Rollout 对象中的 label 的 key
		PodTemplateHashKey string `json:"podTemplateHashKey,omitempty" yaml:"podTemplateHashKey,omitempty"`

		// Revision the desired revision
		Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`

		// Versions versions detail
		// key is pod-template-hash, if equal to PodTemplateHash, the version is the desired version
		Versions map[string]*ClusterVersion `json:"versions,omitempty" yaml:"versions,omitempty"`

		// ManualPaused indicates whether the cluster is in manual pause state
		ManualPaused bool `json:"manualPaused" yaml:"manualPaused"`
	}

	Step struct {
		Index    int   `json:"index"`
		Total    int   `json:"total"`
		Replicas []int `json:"replicas"`
	}

	// ClusterVersion version information
	ClusterVersion struct {
		// Replicas the replicas of this revision
		Replicas int    `json:"replicas,omitempty" yaml:"replicas,omitempty"`
		Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
		// Pods the pods detail of this revision, the key is pod name
		Pods map[string]*ClusterPod `json:"pods,omitempty" yaml:"pods,omitempty"`
	}

	// ClusterPod pod detail
	ClusterPod struct {
		Metadata          PodMetadata  `json:"metadata,omitempty" yaml:"metadata,omitempty"`
		Spec              PodSpec      `json:"spec,omitempty" yaml:"spec,omitempty"`
		Status            PodStatus    `json:"status,omitempty" yaml:"status,omitempty"`
		DeletionTimestamp *metav1.Time `json:"deletionTimestamp,omitempty"`
	}

	PodMetadata struct {
		CreationTimestamp metav1.Time       `json:"creationTimestamp"`
		Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	}

	PodSpec struct {
		NodeName       string       `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
		InitContainers []*Container `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
		Containers     []*Container `json:"containers,omitempty" yaml:"containers,omitempty"`
	}

	Container struct {
		Name  string `json:"name,omitempty" yaml:"name,omitempty"`
		Image string `json:"image,omitempty" yaml:"image,omitempty"`
	}

	PodStatus struct {
		HostIP            string             `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
		PodIP             string             `json:"podIP,omitempty" yaml:"podIP,omitempty"`
		Phase             string             `json:"phase,omitempty" yaml:"phase,omitempty"`
		Events            []Event            `json:"events,omitempty" yaml:"events,omitempty"`
		ContainerStatuses []*ContainerStatus `json:"containerStatuses,omitempty" yaml:"containerStatuses,omitempty"`
		LifeCycle         []*LifeCycleItem   `json:"lifeCycle" yaml:"lifeCycle"`
	}

	ContainerStatus struct {
		Name         string         `json:"name,omitempty" yaml:"name,omitempty"`
		Ready        bool           `json:"ready" yaml:"ready"`
		RestartCount int32          `json:"restartCount"`
		State        ContainerState `json:"state" yaml:"state"`
		ImageID      string         `json:"imageID" yaml:"imageID"`
	}

	Event struct {
		Type           string      `json:"type" yaml:"type"`
		Reason         string      `json:"reason,omitempty" yaml:"reason,omitempty"`
		Message        string      `json:"message,omitempty" yaml:"message,omitempty"`
		Count          int32       `json:"count,omitempty" yaml:"count,omitempty"`
		EventTimestamp metav1.Time `json:"eventTimestamp,omitempty" yaml:"eventTimestamp,omitempty"`
	}

	// ContainerDetail represents more information about a container
	ContainerDetail struct {
		// Name of the container.
		Name string `json:"name"`

		// Image URI of the container.
		Image string `json:"image"`

		// List of environment variables.
		Env []corev1.EnvVar `json:"env"`

		// Commands of the container
		Commands []string `json:"commands"`

		// Command arguments
		Args []string `json:"args"`

		// Information about mounted volumes
		VolumeMounts []VolumeMount `json:"volumeMounts"`

		// Security configuration that will be applied to a container.
		SecurityContext *corev1.SecurityContext `json:"securityContext"`

		// Status of a pod container
		Status *corev1.ContainerStatus `json:"status"`

		// Probes
		LivenessProbe  *corev1.Probe `json:"livenessProbe"`
		ReadinessProbe *corev1.Probe `json:"readinessProbe"`
		StartupProbe   *corev1.Probe `json:"startupProbe"`
	}

	VolumeMount struct {
		// Name of the variable.
		Name string `json:"name"`

		// Is the volume read only ?
		ReadOnly bool `json:"readOnly"`

		// Path within the container at which the volume should be mounted. Must not contain ':'.
		MountPath string `json:"mountPath"`

		// Path within the volume from which the container's volume should be mounted. Defaults to "" (volume's root).
		SubPath string `json:"subPath"`

		// Information about the Volume itself
		Volume corev1.Volume `json:"volume"`
	}

	ContainerState struct {
		State     string       `json:"state" yaml:"state"`
		Reason    string       `json:"reason" yaml:"reason"`
		Message   string       `json:"message" yaml:"message"`
		StartedAt *metav1.Time `json:"startedAt,omitempty"`
	}

	LifeCycleItem struct {
		Type         string      `json:"type" yaml:"type"`
		Status       string      `json:"status" yaml:"status"`
		Message      string      `json:"message" yaml:"status"`
		CompleteTime metav1.Time `json:"completeTime" yaml:"completeTime"`
	}
)

func getPodTemplateHash(rs *appsv1.ReplicaSet) (string, string) {
	podTemplateHash := rs.Labels[DeploymentPodTemplateHash]
	podTemplateHashKey := DeploymentPodTemplateHash
	if podTemplateHash == "" {
		podTemplateHash = rs.Labels[RolloutPodTemplateHash]
		podTemplateHashKey = RolloutPodTemplateHash
	}
	return podTemplateHashKey, podTemplateHash
}

func getRevision(rs *appsv1.ReplicaSet) string {
	revision := rs.Annotations[_deploymentRevision]
	if revision == "" {
		revision = rs.Annotations[_rolloutRevision]
	}
	return revision
}

// CompareRevision
// rs1 版本更加新的话，则返回true，否则返回false
// 注意：仅仅通过 CreationTimestamp 是无法判断最新版本的，尤其是当 集群进行回滚操作。
func CompareRevision(ctx context.Context, rs1, rs2 *appsv1.ReplicaSet) bool {
	revision1 := getRevision(rs1)
	revision2 := getRevision(rs2)

	hasSameOwnerRef := false
	for _, rs1Owner := range rs1.OwnerReferences {
		for _, rs2Owner := range rs2.OwnerReferences {
			if rs1Owner.UID == rs2Owner.UID {
				hasSameOwnerRef = true
				break
			}
		}
	}

	if revision1 == "" || revision2 == "" || !hasSameOwnerRef {
		// 如果它们属于不同的Deployment或Rollout，或
		// 如果某个revision不存在，
		// 则使用 CreationTimestamp 进行比较
		return rs2.CreationTimestamp.Before(&rs1.CreationTimestamp)
	}

	num1, err := strconv.Atoi(revision1)
	if err != nil {
		log.Error(ctx, err)
		return false
	}
	num2, err := strconv.Atoi(revision2)
	if err != nil {
		log.Error(ctx, err)
		return true
	}

	return num1 > num2
}

func getDeploymentCondition(status appsv1.DeploymentStatus,
	condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func getStep(rollout *v1alpha1.Rollout) *Step {
	if rollout == nil {
		return &Step{
			Index:    0,
			Total:    1,
			Replicas: []int{1},
		}
	}

	var replicasTotal = 1
	if rollout.Spec.Replicas != nil {
		replicasTotal = int(*rollout.Spec.Replicas)
	}

	if rollout.Spec.Strategy.Canary == nil ||
		len(rollout.Spec.Strategy.Canary.Steps) == 0 {
		return &Step{
			Index:    0,
			Total:    1,
			Replicas: []int{replicasTotal},
		}
	}

	replicasList := make([]int, 0)
	for _, step := range rollout.Spec.Strategy.Canary.Steps {
		if step.SetWeight != nil {
			replicasList = append(replicasList, int(math.Ceil(float64(*step.SetWeight)/100*float64(replicasTotal))))
		}
	}

	incrementReplicasList := make([]int, 0, len(replicasList))
	for i := 0; i < len(replicasList); i++ {
		replicas := replicasList[i]
		if i > 0 {
			replicas = replicasList[i] - replicasList[i-1]
		}
		incrementReplicasList = append(incrementReplicasList, replicas)
	}

	var stepIndex = 0
	// if steps changes, stepIndex = 0
	if rollout.Status.CurrentStepHash == computeStepHash(rollout) &&
		rollout.Status.CurrentStepIndex != nil {
		index := float64(*rollout.Status.CurrentStepIndex)
		index = math.Min(index, float64(len(rollout.Spec.Strategy.Canary.Steps)))
		for i := 0; i < int(index); i++ {
			if rollout.Spec.Strategy.Canary.Steps[i].SetWeight != nil {
				stepIndex++
			}
		}
	}

	return &Step{
		Index:    stepIndex,
		Total:    len(incrementReplicasList),
		Replicas: incrementReplicasList,
	}
}

func (c *cd) GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error) {
	logStrC := make(chan string)
	logParam := argocd.ContainerLogParams{
		Namespace:     params.Namespace,
		PodName:       params.Pod,
		ContainerName: params.Container,
		TailLines:     params.TailLines,
	}
	argoCD, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	logC, errC, err := argoCD.GetContainerLog(ctx, params.Cluster, logParam)
	if err != nil {
		return nil, err
	}

	go func() {
		for logC != nil || errC != nil {
			select {
			case l, ok := <-logC:
				if !ok {
					logC = nil
					continue
				}
				logStrC <- fmt.Sprintf("[%s] %s\n", l.Result.Timestamp, l.Result.Content)
			case e, ok := <-errC:
				if !ok {
					errC = nil
					continue
				}
				logStrC <- fmt.Sprintf("%s\n", e)
			}
		}
		close(logStrC)
	}()
	return logStrC, nil
}

// onlineCommand the location of online.sh in pod is /home/appops/.probe/online-once.sh
const onlineCommand = `
export ONLINE_SHELL="/home/appops/.probe/online-once.sh"
[[ -f "$ONLINE_SHELL" ]] || {
	echo "there is no online config for this cluster." >&2; exit 1
}

bash "$ONLINE_SHELL"
`

// offlineCommand the location of offline.sh in pod is /home/appops/.probe/offline-once.sh
const offlineCommand = `
export OFFLINE_SHELL="/home/appops/.probe/offline-once.sh"
[[ -f "$OFFLINE_SHELL" ]] || {
	echo "there is no offline config for this cluster." >&2; exit 1
}

bash "$OFFLINE_SHELL"
`

func (c *cd) Online(ctx context.Context, params *ExecParams) (_ map[string]ExecResp, err error) {
	return c.exec(ctx, params, onlineCommand)
}

func (c *cd) Offline(ctx context.Context, params *ExecParams) (_ map[string]ExecResp, err error) {
	return c.exec(ctx, params, offlineCommand)
}

func (c *cd) exec(ctx context.Context, params *ExecParams, command string) (_ map[string]ExecResp, err error) {
	const op = "cd: exec"
	defer wlog.Start(ctx, op).StopPrint()

	config, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}
	containers := make([]kube.ContainerRef, 0)
	for _, pod := range params.PodList {
		containers = append(containers, kube.ContainerRef{
			Config:        config,
			KubeClientset: kubeClient.Basic,
			Namespace:     params.Namespace,
			Pod:           pod,
			Container:     params.Cluster,
		})
	}

	return executeCommandInPods(ctx, containers, []string{"bash", "-c", command}, nil), nil
}

func executeCommandInPods(ctx context.Context, containers []kube.ContainerRef,
	command []string, executor exec.RemoteExecutor) map[string]ExecResp {
	var wg sync.WaitGroup
	ch := make(chan ExecResp, len(containers))
	for _, containerRef := range containers {
		wg.Add(1)
		containerRef := containerRef
		go func(key string) {
			defer wg.Done()
			stdout, stderr, err := kube.Exec(ctx, containerRef, command, executor)
			if err != nil {
				log.Errorf(ctx, "failed to do exec %v, err=%v", command, err)
			}
			ch <- ExecResp{
				key: key,
				Result: func() bool {
					if stderr != "" || err != nil {
						return false
					}
					return true
				}(),
				Stdout: stdout,
				Stderr: stderr,
				Error:  err,
			}
		}(containerRef.Pod)
	}
	wg.Wait()
	close(ch)
	result := make(map[string]ExecResp)
	for val := range ch {
		result[val.key] = val
	}

	return result
}

func getSkipAllStepsPatchStr(stepCnt int) string {
	return fmt.Sprintf(`{"spec":{"paused":false},"status": {"currentStepIndex": %d, "pauseCondition":null}}`,
		stepCnt)
}

func getPausePatchStr() string {
	return `{"spec": {"paused": true}}`
}

func getResumePatchStr() string {
	return `{"spec": {"paused": false}}`
}

// computeStepHash returns a hash value calculated from the Rollout's steps. The hash will
// be safe encoded to avoid bad words.
// source code ref:
// g.hz.netease.com/music-cloud-native/kubernetes/argo-rollouts/-/blob/develop/utils/conditions/conditions.go#L240
func computeStepHash(rollout *v1alpha1.Rollout) string {
	if rollout.Spec.Strategy.BlueGreen != nil || rollout.Spec.Strategy.Canary == nil {
		return ""
	}
	rolloutStepHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(rollout.Spec.Strategy.Canary.Steps)
	if err != nil {
		panic(err)
	}
	_, err = rolloutStepHasher.Write(stepsBytes)
	if err != nil {
		panic(err)
	}
	return rand.SafeEncodeString(fmt.Sprint(rolloutStepHasher.Sum32()))
}
