package cd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"g.hz.netease.com/horizon/pkg/cluster/cd/argocd"
	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	argocdconf "g.hz.netease.com/horizon/pkg/config/argocd"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

const (
	_deploymentRevision       = "deployment.kubernetes.io/revision"
	DeploymentPodTemplateHash = "pod-template-hash"
	_rolloutRevision          = "rollout.argoproj.io/revision"
	RolloutPodTemplateHash    = "rollouts-pod-template-hash"
)

type GetClusterStateParams struct {
	Environment  string
	Cluster      string
	Namespace    string
	RegionEntity *regionmodels.RegionEntity
}

type DeployClusterParams struct {
	Environment   string
	Cluster       string
	GitRepoSSHURL string
	ValueFiles    []string
	RegionEntity  *regionmodels.RegionEntity
	Namespace     string
	Revision      string
}

type CD interface {
	DeployCluster(ctx context.Context, params *DeployClusterParams) error
	// GetClusterState get cluster state in cd system
	GetClusterState(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
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

func (c *cd) DeployCluster(ctx context.Context, params *DeployClusterParams) (err error) {
	const op = "cd: deploy cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return errors.E(op, err)
	}

	// if argo application is not exists, create it first
	if _, err = argo.GetApplication(ctx, params.Cluster); err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return errors.E(op, err)
		}
		var argoApplication = argocd.AssembleArgoApplication(params.Cluster, params.Namespace,
			params.GitRepoSSHURL, params.RegionEntity.K8SCluster.Server, params.ValueFiles)

		manifest, err := json.Marshal(argoApplication)
		if err != nil {
			return errors.E(op, err)
		}

		if err := argo.CreateApplication(ctx, manifest); err != nil {
			return errors.E(op, err)
		}
	}

	return argo.DeployApplication(ctx, params.Cluster, params.Revision)
}

// GetClusterState TODO(gjq) restructure
func (c *cd) GetClusterState(ctx context.Context,
	params *GetClusterStateParams) (clusterState *ClusterState, err error) {
	const op = "cd: get cluster status"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	argo, err := c.factory.GetArgoCD(params.Environment)
	if err != nil {
		return nil, errors.E(op, err)
	}

	_, kubeClient, err := c.kubeClientFty.GetByK8SServer(ctx, params.RegionEntity.K8SCluster.Server)
	if err != nil {
		return nil, errors.E(op, err)
	}

	clusterState = &ClusterState{Versions: map[string]*ClusterVersion{}}

	// get application status
	argoApp, err := argo.GetApplication(ctx, params.Cluster)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// namespace = argoApp.Spec.Destination.Namespace
	clusterState.Status = argoApp.Status.Health.Status
	if clusterState.Status == "" {
		return nil, errors.E(op, http.StatusNotFound, "clusterState.Status == ''")
	}
	if clusterState.Status == health.HealthStatusUnknown {
		clusterState.Status = health.HealthStatusDegraded
	} else if clusterState.Status == health.HealthStatusMissing {
		clusterState.Status = health.HealthStatusProgressing
	}

	var rollout *v1alpha1.Rollout
	if err := argo.GetApplicationResource(ctx, params.Cluster, argocd.ResourceParams{
		Group:        "argoproj.io",
		Version:      "v1alpha1",
		Kind:         "Rollout",
		Namespace:    argoApp.Spec.Destination.Namespace,
		ResourceName: params.Cluster,
	}, &rollout); err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return nil, errors.E(op, err)
		}
	}

	clusterState.Step = getStep(rollout)

	var latestReplicaSet *appsv1.ReplicaSet
	labelSelector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v",
		common.ClusterLabelKey, params.Cluster))
	rss, err := kube.GetReplicaSets(ctx, kubeClient, params.Namespace, labelSelector.String())
	if err != nil {
		return nil, err
	} else if len(rss) == 0 {
		return nil, errors.E(op, http.StatusNotFound, "ReplicaSet instance not found")
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
		return nil, errors.E(op, http.StatusNotFound, "clusterState.PodTemplateHash == ''")
	}

	// 从目前的配置来看，该 if 分支表示负载类型是 serverless 应用
	if clusterState.PodTemplateHashKey == DeploymentPodTemplateHash {
		labelSelector := fields.ParseSelectorOrDie(
			fmt.Sprintf("%v=%v", common.ClusterLabelKey, params.Cluster))
		// serverless 应用会有多个 Deployment 对象
		deploymentList, err := kube.GetDeploymentList(ctx, kubeClient, params.Namespace, labelSelector.String())
		if err != nil {
			return nil, errors.E(op, err)
		}
		for i := range deploymentList {
			deployment := &deploymentList[i]
			// Borrowed at kubernetes/kubectl/rollout_status.go
			if deployment.Generation <= deployment.Status.ObservedGeneration {
				cond := getDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
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

	if err := c.paddingPodAndEventInfo(ctx, params.Cluster, params.Namespace,
		kubeClient, clusterState); err != nil {
		return nil, errors.E(op, err)
	}

	return clusterState, nil
}

func (c *cd) paddingPodAndEventInfo(ctx context.Context, cluster, namespace string,
	kubeClientSet kubernetes.Interface, clusterState *ClusterState) error {
	const op = "deployer: padding pod and event"
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
			return errors.E(op, e)
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
		}
		containerStatuses = append(containerStatuses, c)
	}
	clusterPod.Status.ContainerStatuses = containerStatuses

	for i := range events {
		clusterPod.Status.Events = append(clusterPod.Status.Events,
			Event{
				Reason:         events[i].Reason,
				Message:        events[i].Message,
				Count:          events[i].Count,
				EventTimestamp: events[i].FirstTimestamp,
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

func parseContainerState(containerStatus corev1.ContainerStatus) ContainerState {
	waiting := "waiting"
	running := "running"
	terminated := "terminated"

	// Only one of its members may be specified.
	state := containerStatus.State

	if state.Running != nil {
		return ContainerState{
			State: running,
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
			State:   terminated,
			Reason:  state.Terminated.Reason,
			Message: state.Terminated.Message,
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
	}

	Step struct {
		Index int `json:"index"`
		Total int `json:"total"`
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
		Metadata PodMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
		Spec     PodSpec     `json:"spec,omitempty" yaml:"spec,omitempty"`
		Status   PodStatus   `json:"status,omitempty" yaml:"status,omitempty"`
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
	}

	ContainerStatus struct {
		Name         string         `json:"name,omitempty" yaml:"name,omitempty"`
		Ready        bool           `json:"ready" yaml:"ready"`
		RestartCount int32          `json:"restartCount"`
		State        ContainerState `json:"state" yaml:"state"`
	}

	Event struct {
		Reason         string      `json:"reason,omitempty" yaml:"reason,omitempty"`
		Message        string      `json:"message,omitempty" yaml:"message,omitempty"`
		Count          int32       `json:"count,omitempty" yaml:"count,omitempty"`
		EventTimestamp metav1.Time `json:"eventTimestamp,omitempty" yaml:"eventTimestamp,omitempty"`
	}

	ContainerState struct {
		State   string `json:"state" yaml:"state"`
		Reason  string `json:"reason" yaml:"reason"`
		Message string `json:"message" yaml:"message"`
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
	if rollout == nil || rollout.Spec.Strategy.Canary == nil ||
		len(rollout.Spec.Strategy.Canary.Steps) == 0 {
		return &Step{
			Index: 0,
			Total: 1,
		}
	}
	var stepTotal = 0
	for _, step := range rollout.Spec.Strategy.Canary.Steps {
		if step.SetWeight != nil {
			stepTotal++
		}
	}
	var stepIndex = 0
	if rollout.Status.CurrentStepIndex != nil {
		index := int(*rollout.Status.CurrentStepIndex)
		for i := 0; i < index; i++ {
			if rollout.Spec.Strategy.Canary.Steps[i].SetWeight != nil {
				stepIndex++
			}
		}
	}
	return &Step{
		Index: stepIndex,
		Total: stepTotal,
	}
}
