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
	"net/http"
	"sync"

	applicationV1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutsV1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	kube2 "github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/argocd"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/getter"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	"github.com/horizoncd/horizon/pkg/cluster/kubeclient"
	argocdconf "github.com/horizoncd/horizon/pkg/config/argocd"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"

	gkPattern = "%s/%s"
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

//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/cd/cd_mock.go -package=mock_cd
type CD interface {
	CreateCluster(ctx context.Context, params *CreateClusterParams) error
	DeployCluster(ctx context.Context, params *DeployClusterParams) error
	DeleteCluster(ctx context.Context, params *DeleteClusterParams) error
	Next(ctx context.Context, params *ClusterNextParams) error
	Promote(ctx context.Context, params *ClusterPromoteParams) error
	Pause(ctx context.Context, params *ClusterPauseParams) error
	Resume(ctx context.Context, params *ClusterResumeParams) error
	// Deprecated: GetClusterState get cluster state in cd system
	// replaced by GetClusterStatusV2
	GetClusterState(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
	GetClusterStateV2(ctx context.Context, params *GetClusterStateV2Params) (*ClusterStateV2, error)
	GetResourceTree(ctx context.Context, params *GetResourceTreeParams) ([]ResourceNode, error)
	GetStep(ctx context.Context, params *GetStepParams) (*Step, error)
	GetContainerLog(ctx context.Context, params *GetContainerLogParams) (<-chan string, error)
	GetPod(ctx context.Context, params *GetPodParams) (*corev1.Pod, error)
	GetPodContainers(ctx context.Context, params *GetPodParams) ([]ContainerDetail, error)
	GetPodEvents(ctx context.Context, params *GetPodEventsParams) ([]Event, error)
	Exec(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	// Deprecated
	Online(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	// Deprecated
	Offline(ctx context.Context, params *ExecParams) (map[string]ExecResp, error)
	DeletePods(ctx context.Context, params *DeletePodsParams) (map[string]OperationResult, error)
}

type cd struct {
	kubeClientFty  kubeclient.Factory
	factory        argocd.Factory
	clusterGitRepo gitrepo.ClusterGitRepo
}

func NewCD(clusterGitRepo gitrepo.ClusterGitRepo, argoCDMapper argocdconf.Mapper) CD {
	return &cd{
		kubeClientFty:  kubeclient.Fty,
		factory:        argocd.NewFactory(argoCDMapper),
		clusterGitRepo: clusterGitRepo,
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
	var argoApplication = argo.AssembleArgoApplication(params.Cluster, params.Namespace,
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

func (c *cd) getRollout(ctx context.Context, environment, clusterName string) (*rolloutsV1alpha1.Rollout, error) {
	var rollout *rolloutsV1alpha1.Rollout
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

func (c *cd) GetResourceTree(ctx context.Context,
	params *GetResourceTreeParams) ([]ResourceNode, error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	_, kubeClient, err := c.kubeClientFty.
		GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	// get resourceTreeInArgo
	resourceTreeInArgo, err := argo.GetApplicationTree(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}

	podsMap := make(map[string]*corev1.Pod)
	c.traverseResourceTree(resourceTreeInArgo, func(node *ResourceTreeNode) bool {
		ifContinue := false
		gk := fmt.Sprintf(gkPattern, node.Group, node.Kind)
		workload.LoopAbilities(func(workload workload.Workload) bool {
			if !workload.MatchGK(gk) {
				return true
			}
			gt := getter.New(workload)
			pods, err := gt.ListPods(node.ResourceNode, kubeClient)
			if err != nil {
				return true
			}

			for i := range pods {
				podsMap[string(pods[i].UID)] = &pods[i]
			}
			ifContinue = false
			return false
		})
		return ifContinue
	})

	resourceTree := make([]ResourceNode, 0, len(resourceTreeInArgo.Nodes))
	for _, node := range resourceTreeInArgo.Nodes {
		n := ResourceNode{ResourceNode: node}
		if n.Kind == "Pod" {
			if podDetail, ok := podsMap[n.UID]; ok {
				t := Compact(*podDetail)
				n.PodDetail = &t
			} else {
				continue
			}
		}
		resourceTree = append(resourceTree, n)
	}

	return resourceTree, nil
}

func (c *cd) GetStep(ctx context.Context, params *GetStepParams) (*Step, error) {
	const op = "cd: get step"
	defer wlog.Start(ctx, op).StopPrint()

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	// get resourceTreeInArgo
	resourceTreeInArgo, err := argo.GetApplicationTree(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}

	ifContinue := true
	step := (*workload.Step)(nil)
	c.traverseResourceTree(resourceTreeInArgo, func(node *ResourceTreeNode) bool {
		if !ifContinue {
			return ifContinue
		}
		gk := fmt.Sprintf(gkPattern, node.Group, node.Kind)
		workload.LoopAbilities(func(workload workload.Workload) bool {
			if !workload.MatchGK(gk) {
				return true
			}

			gt := getter.New(workload)
			step, err = gt.GetSteps(node.ResourceNode, kubeClient)
			if err != nil {
				return true
			}

			ifContinue = false
			return false
		})
		return ifContinue
	})

	// step
	if step == nil {
		return &Step{
			Index:        0,
			Total:        0,
			Replicas:     []int{},
			ManualPaused: false,
		}, nil
	}

	return &Step{
		Index:        step.Index,
		Total:        step.Total,
		Replicas:     step.Replicas,
		ManualPaused: step.ManualPaused,
	}, nil
}

// GetClusterStateV2 fetches status of cluster
func (c *cd) GetClusterStateV2(ctx context.Context,
	params *GetClusterStateV2Params) (*ClusterStateV2, error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	// get application status
	argoApp, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}

	if argoApp.Status.Health.Status == "" {
		return nil, perror.Wrapf(
			herrors.NewErrNotFound(herrors.ClusterStateInArgo, "cluster not found in argo"),
			"failed to get cluster status from argo: app name = %v", params.Cluster)
	}

	status := &ClusterStateV2{
		Status: string(argoApp.Status.Health.Status),
	}

	if status.Status != string(health.HealthStatusHealthy) {
		return status, nil
	}

	if argoApp.Status.Sync.Status != applicationV1alpha1.SyncStatusCodeSynced {
		status.Status = string(health.HealthStatusProgressing)
		return status, nil
	}

	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, params.Application, params.Cluster)
	if err != nil {
		return nil, err
	}
	if lastConfigCommit.Master != argoApp.Status.Sync.Revision {
		status.Status = string(health.HealthStatusProgressing)
		log.Warningf(ctx,
			"current revision(%s) is not consistent with gitops repo commit(%s)",
			argoApp.Status.Sync.Revision, lastConfigCommit.Master)
		return status, nil
	}

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	// get resourceTreeInArgo
	resourceTreeInArgo, err := argo.GetApplicationTree(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}

	if argoApp.Status.Health.Status == health.HealthStatusHealthy {
		isHealthy := true
		c.traverseResourceTree(resourceTreeInArgo, func(node *ResourceTreeNode) bool {
			if !isHealthy {
				return false
			}
			gk := fmt.Sprintf(gkPattern, node.Group, node.Kind)
			workload.LoopAbilities(func(workload workload.Workload) bool {
				if !workload.MatchGK(gk) {
					return true
				}
				gt := getter.New(workload)
				nodeHealthy, err := gt.IsHealthy(node.ResourceNode, kubeClient)
				if err != nil {
					return true
				}
				log.Debugf(ctx, "[cd get status v2] node(%v) kind(%v) isHealthy(%v)", node.Name, node.Kind, nodeHealthy)
				isHealthy = isHealthy && nodeHealthy
				return isHealthy
			})
			// break if isHealthy is false
			return isHealthy
		})

		if !isHealthy {
			status.Status = string(health.HealthStatusProgressing)
		}
	}
	return status, nil
}

// Deprecated: GetClusterState
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

	// TODO: rollout coupling
	var rollout *rolloutsV1alpha1.Rollout
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterClusterLabelKey, params.Cluster))
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
						clusterPodMap[node.Name] = podMapping(podMap[node.Name])
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

	if clusterState.PodTemplateHashKey == DeploymentPodTemplateHash {
		labelSelector := fields.ParseSelectorOrDie(
			fmt.Sprintf("%v=%v", common.ClusterClusterLabelKey, params.Cluster))
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
					// By default, if a Deployment fails to complete a rollover within ten minutes,
					// Then the Deployment's health status is HealthStatusDegraded.
					clusterState.Status = health.HealthStatusDegraded
				}
			} else {
				// If the Deployment has an update that has not been processed by the Deployment Controller,，
				// Then the Deployment's health status is HealthStatusProgressing.
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
		common.ClusterClusterLabelKey, params.Cluster))
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

// Deprecated
func (c *cd) paddingPodAndEventInfo(ctx context.Context, cluster, namespace string,
	kubeClient kubernetes.Interface, clusterState *ClusterState) error {
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterClusterLabelKey, cluster))

	var pods []corev1.Pod
	var events map[string][]*corev1.Event
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pods, err1 = kube.GetPods(ctx, kubeClient, namespace, labelSelector.String())
	}()
	go func() {
		defer wg.Done()
		events, err2 = kube.GetEvents(ctx, kubeClient, namespace)
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
			log.Info(ctx, err)
			continue
		} else {
			clusterState.Replicas++
		}
	}
	return nil
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

// Deprecated
func (c *cd) Online(ctx context.Context, params *ExecParams) (_ map[string]ExecResp, err error) {
	return c.exec(ctx, params, onlineCommand)
}

// Deprecated
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

func (c *cd) Exec(ctx context.Context, params *ExecParams) (_ map[string]ExecResp, err error) {
	const op = "cd: shell exec"
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

	return executeCommandInPods(ctx, containers, params.Commands, nil), nil
}

// TraverseOperator stops if result is false
type TraverseOperator func(node *ResourceTreeNode) bool

// traverseResourceTree traverses tree by dfs
func (c *cd) traverseResourceTree(resourceTree *applicationV1alpha1.ApplicationTree,
	operators ...TraverseOperator) {
	m := make(map[string]*applicationV1alpha1.ResourceNode)
	for i, node := range resourceTree.Nodes {
		m[node.UID] = &resourceTree.Nodes[i]
	}

	visited := make(map[string]*ResourceTreeNode)
	roots := make([]*ResourceTreeNode, 0, 4)
	for i := range resourceTree.Nodes {
		tree := (*ResourceTreeNode)(nil)
		currentNode := &resourceTree.Nodes[i]
		for {
			if _, ok := visited[currentNode.UID]; ok {
				parent := visited[currentNode.UID]
				if tree != nil {
					parent.children = append(parent.children, tree)
				}
				break
			}

			t := &ResourceTreeNode{
				ResourceNode: currentNode,
			}
			if tree != nil {
				t.children = append(t.children, tree)
			}
			tree = t
			visited[currentNode.UID] = tree

			if currentNode.ParentRefs != nil {
				currentNode = m[currentNode.ParentRefs[0].UID]
			} else {
				roots = append(roots, tree)
				break
			}
		}
	}

	var dfs func(node *ResourceTreeNode, operator TraverseOperator)
	dfs = func(node *ResourceTreeNode, operator TraverseOperator) {
		if node == nil {
			return
		}
		if !operator(node) || node.children == nil {
			return
		}
		for _, child := range node.children {
			dfs(child, operator)
		}
	}

	for _, operator := range operators {
		for _, root := range roots {
			dfs(root, operator)
		}
	}
}
