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
	kubeutil "github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/argocd"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	"github.com/horizoncd/horizon/pkg/cluster/kubeclient"
	argocdconf "github.com/horizoncd/horizon/pkg/config/argocd"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/regioninformers"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/horizoncd/horizon/pkg/workload"
	"github.com/horizoncd/horizon/pkg/workload/getter"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"
)

var (
	GKPod = schema.GroupKind{
		Group: "",
		Kind:  "Pod",
	}
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

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/cd/cd_mock.go -package=mock_cd
type CD interface {
	CreateCluster(ctx context.Context, params *CreateClusterParams) error
	DeployCluster(ctx context.Context, params *DeployClusterParams) error
	DeleteCluster(ctx context.Context, params *DeleteClusterParams) error
	GetClusterState(ctx context.Context, params *GetClusterStateV2Params) (*ClusterStateV2, error)
	GetResourceTree(ctx context.Context, params *GetResourceTreeParams) ([]ResourceNode, error)
	GetStep(ctx context.Context, params *GetStepParams) (*Step, error)
	GetPodEvents(ctx context.Context, params *GetPodEventsParams) ([]Event, error)
}

type cd struct {
	kubeClientFactory kubeclient.Factory
	informerFactories *regioninformers.RegionInformers
	factory           argocd.Factory
	clusterGitRepo    gitrepo.ClusterGitRepo
	targetRevision    string
}

func NewCD(informerFactories *regioninformers.RegionInformers, clusterGitRepo gitrepo.ClusterGitRepo,
	argoCDMapper argocdconf.Mapper, targetRevision string) CD {
	return &cd{
		kubeClientFactory: kubeclient.Fty,
		informerFactories: informerFactories,
		factory:           argocd.NewFactory(argoCDMapper),
		clusterGitRepo:    clusterGitRepo,
		targetRevision:    targetRevision,
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
		params.GitRepoURL, params.RegionEntity.Server, params.ValueFiles, c.targetRevision)

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

func (c *cd) GetResourceTree(ctx context.Context,
	params *GetResourceTreeParams) ([]ResourceNode, error) {
	const op = "cd: get resource tree"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	// get resourceTreeInArgo
	resourceTreeInArgo, err := argo.GetApplicationTree(ctx, params.Cluster)
	if err != nil {
		return nil, err
	}

	resourceTree := make([]ResourceNode, 0, len(resourceTreeInArgo.Nodes))
	pd, err := workload.GetAbility(GKPod)
	if err != nil {
		return nil, err
	}
	gt := getter.New(pd)
	for _, node := range resourceTreeInArgo.Nodes {
		n := ResourceNode{ResourceNode: node}
		if n.Kind == "Pod" {
			var podDetail corev1.Pod
			err = c.informerFactories.GetDynamicFactory(params.RegionEntity.ID,
				func(factory dynamicinformer.DynamicSharedInformerFactory) error {
					log.Debugf(ctx, "get pod detail: %v", node.Name)
					pods, err := gt.ListPods(&node, factory)
					if err != nil {
						log.Debugf(ctx, "failed to get pod detail: %v", err)
						return err
					}
					log.Debugf(ctx, "get pod detail success: %v", node.Name)
					podDetail = pods[0]
					return nil
				})
			if err != nil {
				log.Errorf(ctx, "failed to get pod detail: %v", err)
				continue
			}
			t := Compact(podDetail)
			n.PodDetail = &t
		}
		resourceTree = append(resourceTree, n)
	}

	return resourceTree, nil
}

func (c *cd) GetStep(ctx context.Context, params *GetStepParams) (*Step, error) {
	const op = "cd: get step"
	defer wlog.Start(ctx, op).StopPrint()

	_, kubeClient, err := c.kubeClientFactory.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
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
		workload.LoopAbilities(func(workload workload.Workload) bool {
			if !workload.MatchGK(schema.GroupKind{Group: node.Group, Kind: node.Kind}) {
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
			AutoPromote:  false,
		}, nil
	}

	return &Step{
		Index:        step.Index,
		Total:        step.Total,
		Replicas:     step.Replicas,
		ManualPaused: step.ManualPaused,
		AutoPromote:  step.AutoPromote,
		Extra:        step.Extra,
	}, nil
}

// GetClusterState fetches status of cluster
func (c *cd) GetClusterState(ctx context.Context,
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

	_, kubeClient, err := c.kubeClientFactory.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
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
			workload.LoopAbilities(func(workload workload.Workload) bool {
				if !workload.MatchGK(schema.GroupKind{Group: node.Group, Kind: node.Kind}) {
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

// Deprecated: using GetClusterState instead
func (c *cd) GetClusterStateV1(ctx context.Context,
	params *GetClusterStateParams) (clusterState *ClusterState, err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, err
	}

	_, kubeClient, err := c.kubeClientFactory.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
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
			if !resourceTreeContains(resourceTree, kubeutil.DeploymentKind) {
				allPods, err := kube.GetPods(ctx, kubeClient.Basic, namespace, labelSelector.String())
				if err != nil {
					return nil, err
				}
				for _, pod := range allPods {
					podMap[pod.Name] = pod
				}
				for _, node := range resourceTree.Nodes {
					if node.Kind == kubeutil.PodKind {
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

func (c *cd) GetPodEvents(ctx context.Context,
	params *GetPodEventsParams) (events []Event, err error) {
	const op = "cd: get cluster pod events"
	defer wlog.Start(ctx, op).StopPrint()

	_, kubeClient, err := c.kubeClientFactory.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	resourceTree, err := c.GetResourceTree(ctx, &GetResourceTreeParams{
		Environment:  params.Environment,
		Cluster:      params.Cluster,
		RegionEntity: params.RegionEntity,
	})
	if err != nil {
		return nil, err
	}

	for i := range resourceTree {
		pod := resourceTree[i].PodDetail
		if pod != nil && pod.Metadata.Namespace == params.Namespace && pod.Metadata.Name == params.Pod {
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
