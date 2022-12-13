package kservice

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	servicev1 "knative.dev/serving/pkg/apis/serving/v1"
)

var Ability = &service{}

type service struct{}

func (s *service) GetRevisions(un *unstructured.Unstructured,
	resourceTree map[string]*v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*workload.Revision, error) {
	var ksvc servicev1.Service
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &ksvc)
	if err != nil {
		return "", nil, err
	}

	selector := polymorphichelpers.MakeLabels(ksvc.Spec.Template.Labels)
	pods, err := client.Basic.CoreV1().Pods(ksvc.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return "", nil, err
	}

	unRevisions, err := client.Dynamic.Resource(schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "revisions",
	}).Namespace(ksvc.Namespace).
		List(context.TODO(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", ksvc.Name)})
	if err != nil {
		return "", nil, err
	}
	var serviceRevisions servicev1.RevisionList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unRevisions.UnstructuredContent(), &serviceRevisions)
	if err != nil {
		return "", nil, err
	}

	revisions := make(map[string]*workload.Revision)
	for _, revision := range serviceRevisions.Items {
		revisions[revision.Name] = &workload.Revision{Name: revision.Name}
	}

	for _, pod := range pods.Items {
		revisionName := pod.Labels["serving.knative.dev/revision"]
		revision := revisions[revisionName]
		if revision != nil {
			revision.Pods = append(revision.Pods, pod)
		}
	}
	return ksvc.Status.LatestCreatedRevisionName, revisions, nil
}

func (*service) ListPods(un *unstructured.Unstructured,
	resourceTree []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	var ksvc servicev1.Service
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &ksvc)
	if err != nil {
		return nil, err
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: ksvc.Spec.Template.Labels})
	pods, err := client.Basic.CoreV1().Pods(ksvc.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}
