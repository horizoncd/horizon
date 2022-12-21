package kservice

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	servicev1 "knative.dev/serving/pkg/apis/serving/v1"
)

func init() {
	workload.Register(Ability, 0)
}

// please refer to github.com/horizoncd/horizon/pkg/cluster/cd/workload/workload.go
var Ability = &service{}

type service struct{}

func (*service) MatchGK(gk string) bool {
	return "serving.knative.dev/Service" == gk
}

func (*service) getServiceByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*servicev1.Service, error) {
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  node.Version,
		Resource: "services",
	}

	un, err := client.Dynamic.Resource(gvr).Namespace(node.Namespace).
		Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get deployment in k8s"),
			"failed to get deployment in k8s: deployment = %s, err = %v", node.Name, err)
	}

	var ksvc *servicev1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &ksvc)
	if err != nil {
		return nil, err
	}
	return ksvc, nil
}

func (s *service) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	instance, err := s.getServiceByNode(node, client)
	if err != nil {
		return nil, err
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: instance.Spec.Template.Labels})
	pods, err := client.Basic.CoreV1().Pods(instance.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}
