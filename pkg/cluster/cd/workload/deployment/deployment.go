package deployment

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	"github.com/horizoncd/horizon/pkg/util/kube"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	deployutil "k8s.io/kubernetes/pkg/controller/deployment/util"
)

var Ability = &deployment{}

type deployment struct {
}

func (*deployment) IsHealthy(un *unstructured.Unstructured,
	client *kube.Client) (bool, error) {
	var deploy v1.Deployment

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &deploy)
	if err != nil {
		return true, err
	}

	if deploy.Status.ObservedGeneration != deploy.Generation {
		return false, nil
	}

	return deploy.Status.AvailableReplicas == *deploy.Spec.Replicas, nil
}

func (*deployment) ListPods(un *unstructured.Unstructured,
	_ []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	var deploy v1.Deployment

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &deploy)
	if err != nil {
		return nil, err
	}

	selector := deploy.Spec.Selector
	pods, err := client.Basic.CoreV1().Pods(deploy.ObjectMeta.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: polymorphichelpers.MakeLabels(selector.MatchLabels),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (*deployment) GetRevisions(un *unstructured.Unstructured,
	_ []v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*workload.Revision, error) {
	var deploy v1.Deployment

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &deploy)
	if err != nil {
		return "", nil, err
	}
	_, allRSs, newRs, err := deployutil.GetAllReplicaSets(&deploy, client.Basic.AppsV1())
	if err != nil {
		return "", nil, err
	}

	currentRevision, err := getRevision(newRs)
	if err != nil {
		return "", nil, err
	}
	allRSs = append(allRSs, newRs)

	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return "", nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	allPods, err := client.Basic.CoreV1().Pods(deploy.Namespace).List(context.TODO(), options)
	if err != nil {
		return "", nil, err
	}

	revisions := make(map[string]*workload.Revision)
	for _, RSs := range allRSs {
		if v, err := getRevision(RSs); err != nil {
			continue
		} else {
			revision := &workload.Revision{Name: v}
			revisions[string(RSs.UID)] = revision
		}
	}

	for i := range allPods.Items {
		pod := &allPods.Items[i]
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef != nil && revisions[string(controllerRef.UID)] != nil {
			revision := revisions[string(controllerRef.UID)]
			revision.Pods = append(revision.Pods, *pod)
		}
	}

	res := make(map[string]*workload.Revision)
	for _, revision := range revisions {
		if len(revision.Pods) != 0 {
			res[revision.Name] = revision
		}
	}

	return currentRevision, res, nil
}

func getRevision(obj runtime.Object) (string, error) {
	v, err := deployutil.Revision(obj)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Rev:%d", v), nil
}
