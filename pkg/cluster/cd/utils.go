package cd

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	applicationV1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutsV1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/exec"
)

func resourceTreeContains(resourceTree *applicationV1alpha1.ApplicationTree, resourceKind string) bool {
	for _, node := range resourceTree.Nodes {
		if node.Kind == resourceKind {
			return true
		}
	}
	return false
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

	cs := &containerList{name: pod.Labels[common.ClusterClusterLabelKey]}
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

// Deprecated: allContainersStarted determine if all containers have been started
func allContainersStarted(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Started == nil || !*(containerStatus.Started) {
			return false
		}
	}
	return true
}

// Deprecated: allContainersRunning determine if all containers running
func allContainersRunning(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.State.Running == nil {
			return false
		}
	}
	return true
}

// Deprecated: allContainersReady determine if all containers ready
func allContainersReady(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return true
}

// Deprecated: oneOfContainersCrash determine if one of containers crash
func oneOfContainersCrash(containerStatuses []corev1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == PodErrCrashLoopBackOff {
			return true
		}
	}
	return false
}

// Deprecated
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

func getStep(rollout *rolloutsV1alpha1.Rollout) *Step {
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
	if rollout.Status.CurrentStepHash == computeRolloutStepHash(rollout) &&
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

// computeRolloutStepHash returns a hash value calculated from the Rollout's steps. The hash will
// be safe encoded to avoid bad words.
// source code ref:
// g.hz.netease.com/music-cloud-native/kubernetes/argo-rollouts/-/blob/develop/utils/conditions/conditions.go#L240
func computeRolloutStepHash(rollout *rolloutsV1alpha1.Rollout) string {
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

func podMapping(pod corev1.Pod) *ClusterPod {
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
	return clusterPod
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

func mapKind2Resource(group, version, kind string, kubeClient kubernetes.Interface) (string, error) {
	apiList, err := kubeClient.Discovery().ServerResourcesForGroupVersion(fmt.Sprintf("%s/%s", group, version))
	if err != nil {
		return "", perror.Wrapf(herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to get api list"),
			"failed to get resource api list: gvk = %s/%s/%s, err = %v", group, version, kind, err)
	}
	for _, resource := range apiList.APIResources {
		if resource.Kind == kind {
			return resource.Name, nil
		}
	}

	// default
	return fmt.Sprintf("%ss", strings.ToLower(kind)), nil
}

// findTop finds the one who has the most pods
func findTop(nodeTree map[string]*applicationV1alpha1.ResourceNode) *applicationV1alpha1.ResourceNode {
	type node struct {
		*applicationV1alpha1.ResourceNode
		count int
	}
	var maxDepth = 100

	m := make(map[string]*node, len(nodeTree))

	for k, v := range nodeTree {
		if v.Kind == "Pod" {
			m[k] = &node{v, 1}
		} else {
			m[k] = &node{v, 0}
		}
	}

	var findParent func(string, int) *node

	findParent = func(uid string, depth int) *node {
		if depth > maxDepth {
			return nil
		}
		n := m[uid]
		if n == nil {
			return nil
		}
		if len(n.ParentRefs) == 0 {
			return n
		}
		return findParent(n.ParentRefs[0].UID, depth+1)
	}

	var maxNode *node
	for k := range m {
		parent := findParent(k, 0)
		if parent != nil {
			parent.count += m[k].count
			if maxNode == nil {
				maxNode = parent
			}
			if maxNode.count < parent.count {
				maxNode = parent
			}
		}
	}

	if maxNode == nil {
		return nil
	}
	return maxNode.ResourceNode
}
