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

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/argocd"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/ability"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/deployment"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/generic"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/kservice"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload/rollout"
	"github.com/horizoncd/horizon/pkg/cluster/kubeclient"
	argocdconf "github.com/horizoncd/horizon/pkg/config/argocd"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"k8s.io/client-go/kubernetes"

	applicationV1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutsV1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	kube2 "github.com/argoproj/gitops-engine/pkg/utils/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/describe"
	kubectlresource "k8s.io/kubectl/pkg/util/resource"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"

	gvkPattern = "%s/%s/%s"

	workloadRollout    = "argoproj.io/v1alpha1/Rollout"
	workloadDeployment = "apps/v1/Deployment"
	workloadKsvc       = "serving.knative.dev/v1/Service"
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
	// GetClusterState get cluster state in cd system
	GetClusterState(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
	GetClusterStateV2(ctx context.Context, params *GetClusterStateV2Params) (*ClusterStateV2, error)
	GetResourceTree(ctx context.Context, params *GetResourceTreeParams) ([]ResourceNode, error)
	GetStep(ctx context.Context, params *GetStepParams) (*Step, error)
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

	un, workloadType, err := c.getTopWorkload(ctx, params.Cluster, params.Environment, params.RegionEntity)
	if err != nil {
		return nil, err
	}

	getter := ability.New(workloadType)

	// TODO: delete this after using informer
	// add detail for pods
	pods, err := getter.ListPods(un, resourceTreeInArgo.Nodes, kubeClient)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*CompactPod, len(pods))
	for i := range pods {
		t := Compact(pods[i])
		m[string(pods[i].UID)] = &t
	}

	resourceTree := make([]ResourceNode, 0, len(resourceTreeInArgo.Nodes))
	for _, node := range resourceTreeInArgo.Nodes {
		n := ResourceNode{ResourceNode: node}
		if n.Kind == "Pod" {
			if podDetail, ok := m[n.UID]; ok {
				n.PodDetail = podDetail
			} else {
				continue
			}
		}
		resourceTree = append(resourceTree, n)
	}

	return resourceTree, nil
}

func (c *cd) getTopWorkload(ctx context.Context, clusterName, environment string,
	region *regionmodels.RegionEntity) (un *unstructured.Unstructured, workloadType interface{}, err error) {
	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(region.Server, region.Certificate)
	if err != nil {
		return
	}

	argo, err := c.factory.GetArgoCD(environment)
	if err != nil {
		return
	}

	// get application status
	argoApp, err := argo.GetApplication(ctx, clusterName)
	if err != nil {
		return
	}
	namespace := argoApp.Spec.Destination.Namespace

	resourceTree, err := argo.GetApplicationTree(ctx, clusterName)
	if err != nil {
		return
	}

	resourceMap := make(map[string]*applicationV1alpha1.ResourceNode)
	gvkMap := make(map[string][]*applicationV1alpha1.ResourceNode)
	for i := range resourceTree.Nodes {
		resourceNode := resourceTree.Nodes[i]
		resourceMap[resourceNode.UID] = &resourceNode
		key := fmt.Sprintf(gvkPattern, resourceNode.Group, resourceNode.Version, resourceNode.Kind)
		gvkMap[key] = append(gvkMap[key], &resourceNode)
	}

	var (
		gvr  schema.GroupVersionResource
		node *applicationV1alpha1.ResourceNode
	)
	if nodes, ok := gvkMap[workloadRollout]; ok {
		node = nodes[0]
		workloadType = rollout.Ability
		gvr = schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "rollouts"}
	} else if nodes, ok = gvkMap[workloadKsvc]; ok {
		node = nodes[0]
		workloadType = kservice.Ability
		gvr = schema.GroupVersionResource{Group: "serving.knative.dev", Version: "v1", Resource: "services"}
	} else if nodes, ok = gvkMap[workloadDeployment]; ok {
		node = nodes[0]
		workloadType = deployment.Ability
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	} else {
		workloadType = generic.Ability
		node = findTop(resourceMap)
		if node == nil {
			return nil, nil, perror.Wrapf(herrors.ErrTopResourceNotFound,
				"failed to get top resource: resource tree = %v", resourceMap)
		}
		resource, err := mapKind2Resource(node.Group, node.Version, node.Kind, kubeClient.Basic)
		if err != nil {
			return nil, nil, err
		}
		gvr = schema.GroupVersionResource{
			Group: node.Group, Version: node.Version, Resource: resource,
		}
	}

	un, err = kubeClient.Dynamic.Resource(gvr).
		Namespace(namespace).Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to get resource"),
			"failed to get resource on k8s: gvr = %v, err = %v", gvr, err)
	}
	return
}

func (c *cd) GetStep(ctx context.Context, params *GetStepParams) (*Step, error) {
	const op = "cd: get step"
	defer wlog.Start(ctx, op).StopPrint()

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(params.RegionEntity.Server, params.RegionEntity.Certificate)
	if err != nil {
		return nil, err
	}

	un, workloadType, err := c.getTopWorkload(ctx,
		params.Cluster, params.Environment, params.RegionEntity)
	if err != nil {
		return nil, err
	}

	getter := ability.New(workloadType)

	// step
	step, err := getter.GetSteps(un, kubeClient)
	if err != nil {
		return nil, err
	}

	return &Step{
		Index:        step.Index,
		Total:        step.Total,
		Replicas:     step.Replicas,
		ManualPaused: step.ManualPaused,
	}, nil
}

// GetClusterStatusV2 fetchs status of cluster
func (c *cd) GetClusterStateV2(ctx context.Context,
	params *GetClusterStateV2Params) (status *ClusterStateV2, err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).StopPrint()

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return
	}

	// get application status
	argoApp, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		return
	}

	if argoApp.Status.Health.Status == "" {
		return status, perror.Wrapf(
			herrors.NewErrNotFound(herrors.ClusterStateInArgo, "cluster not found in argo"),
			"failed to get cluster status from argo: app name = %v", params.Cluster)
	}

	// TODO: check commit revision
	status = &ClusterStateV2{string(argoApp.Status.Health.Status)}
	if argoApp.Status.Sync.Status != applicationV1alpha1.SyncStatusCodeSynced {
		status.Status = string(health.HealthStatusProgressing)
	}
	return
}

// Deprecated: GetClusterState TODO(gjq) restructure
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
