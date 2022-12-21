package rollout

import (
	"context"
	"math"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutsv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	workload.Register(ability)
}

// please refer to github.com/horizoncd/horizon/pkg/cluster/cd/workload/workload.go
var ability = &rollout{}

type rollout struct {
}

func (*rollout) MatchGK(gk string) bool {
	return "argoproj.io/Rollout" == gk
}

func (*rollout) getRolloutByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*rolloutsv1alpha1.Rollout, error) {
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  node.Version,
		Resource: "rollouts",
	}

	un, err := client.Dynamic.Resource(gvr).Namespace(node.Namespace).
		Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get rollout in k8s"),
			"failed to get rollout in k8s: deployment = %s, err = %v", node.Name, err)
	}

	var instance *rolloutsv1alpha1.Rollout
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (r *rollout) IsHealthy(node *v1alpha1.ResourceNode,
	client *kube.Client) (bool, error) {
	instance, err := r.getRolloutByNode(node, client)
	if err != nil {
		return true, err
	}

	if instance == nil {
		return true, nil
	}

	labels := instance.Status.Selector
	pods, err := client.Basic.CoreV1().Pods(instance.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return true, err
	}

	count := 0
	required := int(*instance.Spec.Replicas)

	templateHashSum := ComputePodSpecHash(instance.Spec.Template.Spec)
	for _, pod := range pods.Items {
		hashSum := ComputePodSpecHash(pod.Spec)
		if templateHashSum != hashSum {
			continue
		}
		for k, v := range instance.Spec.Template.ObjectMeta.Annotations {
			if pod.Annotations[k] != v {
				continue
			}
		}
		for k, v := range instance.Spec.Template.ObjectMeta.Labels {
			if pod.Annotations[k] != v {
				continue
			}
		}
		count++
	}
	if count != required {
		return false, nil
	}

	return int(*instance.Status.CurrentStepIndex) == len(instance.Spec.Strategy.Canary.Steps), nil
}

func (r *rollout) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	instance, err := r.getRolloutByNode(node, client)
	if err != nil {
		return nil, err
	}

	pods, err := client.Basic.CoreV1().Pods(instance.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: instance.Status.Selector})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (r *rollout) GetSteps(node *v1alpha1.ResourceNode, client *kube.Client) (*workload.Step, error) {
	instance, err := r.getRolloutByNode(node, client)
	if err != nil {
		return nil, err
	}

	var replicasTotal = 1
	if instance.Spec.Replicas != nil {
		replicasTotal = int(*instance.Spec.Replicas)
	}

	if instance.Spec.Strategy.Canary == nil ||
		len(instance.Spec.Strategy.Canary.Steps) == 0 {
		return &workload.Step{
			Index:    0,
			Total:    1,
			Replicas: []int{replicasTotal},
		}, nil
	}

	replicasList := make([]int, 0)
	for _, step := range instance.Spec.Strategy.Canary.Steps {
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
	if instance.Status.CurrentStepHash == computeStepHash(instance) &&
		instance.Status.CurrentStepIndex != nil {
		index := float64(*instance.Status.CurrentStepIndex)
		index = math.Min(index, float64(len(instance.Spec.Strategy.Canary.Steps)))
		for i := 0; i < int(index); i++ {
			if instance.Spec.Strategy.Canary.Steps[i].SetWeight != nil {
				stepIndex++
			}
		}
	}

	// manual paused
	return &workload.Step{
		Index:        stepIndex,
		Total:        len(incrementReplicasList),
		Replicas:     incrementReplicasList,
		ManualPaused: instance.Spec.Paused,
	}, nil
}
