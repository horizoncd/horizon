package rollout

import (
	"context"
	"math"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutsv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	revisionKey = "Revision"
)

var Ability = &rollout{}

type rollout struct {
}

func fetchInfo(infos []v1alpha1.InfoItem, key string) string {
	for _, info := range infos {
		if info.Name == key {
			return info.Value
		}
	}
	return ""
}

func (*rollout) IsHealthy(un *unstructured.Unstructured,
	client *kube.Client) (bool, error) {
	var rollout *rolloutsv1alpha1.Rollout
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &rollout)
	if err != nil {
		return true, err
	}

	if rollout == nil {
		return true, nil
	}

	labels := rollout.Status.Selector
	pods, err := client.Basic.CoreV1().Pods(rollout.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return true, err
	}

	count := 0
	required := int(*rollout.Spec.Replicas)

	templateHashSum := ComputePodSpecHash(rollout.Spec.Template.Spec)
	for _, pod := range pods.Items {
		hashSum := ComputePodSpecHash(pod.Spec)
		if templateHashSum != hashSum {
			continue
		}
		for k, v := range rollout.Spec.Template.ObjectMeta.Annotations {
			if pod.Annotations[k] != v {
				continue
			}
		}
		for k, v := range rollout.Spec.Template.ObjectMeta.Labels {
			if pod.Annotations[k] != v {
				continue
			}
		}
		count++
	}
	if count != required {
		return false, nil
	}

	return int(*rollout.Status.CurrentStepIndex) == len(rollout.Spec.Strategy.Canary.Steps), nil
}

func (*rollout) ListPods(un *unstructured.Unstructured,
	resourceTree []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	var rollout *rolloutsv1alpha1.Rollout
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &rollout)
	if err != nil {
		return nil, err
	}

	pods, err := client.Basic.CoreV1().Pods(rollout.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: rollout.Status.Selector})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (*rollout) GetRevisions(un *unstructured.Unstructured,
	resourceTree []v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*workload.Revision, error) {
	var rollout *rolloutsv1alpha1.Rollout
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &rollout)
	if err != nil {
		return "", nil, err
	}

	if rollout == nil {
		return "", nil, nil
	}

	labels := rollout.Status.Selector
	pods, err := client.Basic.CoreV1().Pods(rollout.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return "", nil, err
	}
	podsMap := make(map[string]corev1.Pod)
	for _, pod := range pods.Items {
		podsMap[string(pod.UID)] = pod
	}

	m := make(map[string]*v1alpha1.ResourceNode)
	for _, node := range resourceTree {
		m[node.UID] = &node
	}

	revisions := make(map[*v1alpha1.ResourceNode]*workload.Revision)

	rolloutNode := m[string(rollout.UID)]
	currentRevision := ""
	if rolloutNode != nil {
		currentRevision = fetchInfo(rolloutNode.Info, revisionKey)
	}

	for _, node := range resourceTree {
		if node.Kind == "Pod" {
			if len(node.ParentRefs) > 0 {
				parentsID := node.ParentRefs[0].UID
				rsNode := m[parentsID]
				if revision, ok := revisions[rsNode]; ok {
					revision.Pods = append(revision.Pods, podsMap[node.UID])
				} else {
					revName := fetchInfo(rsNode.Info, revisionKey)
					if revName == "" {
						revName = rsNode.Name
					}
					if currentRevision == "" {
						currentRevision = revName
					}
					revision := workload.Revision{
						Name: revName,
						Pods: []corev1.Pod{podsMap[node.UID]},
					}
					revisions[rsNode] = &revision
				}
			}
		}
	}

	res := make(map[string]*workload.Revision)
	for _, revision := range revisions {
		res[revision.Name] = revision
	}

	return currentRevision, res, nil
}

func (*rollout) GetSteps(un *unstructured.Unstructured, _ *kube.Client) (*workload.Step, error) {
	var rollout *rolloutsv1alpha1.Rollout
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &rollout)
	if err != nil {
		return nil, err
	}

	var replicasTotal = 1
	if rollout.Spec.Replicas != nil {
		replicasTotal = int(*rollout.Spec.Replicas)
	}

	if rollout.Spec.Strategy.Canary == nil ||
		len(rollout.Spec.Strategy.Canary.Steps) == 0 {
		return &workload.Step{
			Index:    0,
			Total:    1,
			Replicas: []int{replicasTotal},
		}, nil
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

	// manual paused
	return &workload.Step{
		Index:        stepIndex,
		Total:        len(incrementReplicasList),
		Replicas:     incrementReplicasList,
		ManualPaused: rollout.Spec.Paused,
	}, nil
}
