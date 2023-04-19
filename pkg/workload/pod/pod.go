package pod

import (
	"context"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	workload.Register(ability)
}

var ability = &pod{}

type pod struct{}

func (*pod) MatchGK(gk schema.GroupKind) bool {
	return gk.Kind == "Pod"
}

func (*pod) getPodByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*corev1.Pod, error) {
	instance, err := client.Basic.CoreV1().Pods(node.Namespace).Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get deployment in k8s"),
			"failed to get deployment in k8s: deployment = %s, err = %v", node.Name, err)
	}
	return instance, nil
}

func (p *pod) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	instance, err := p.getPodByNode(node, client)
	if err != nil {
		return nil, err
	}

	return []corev1.Pod{*instance}, nil
}

func (*pod) Action(actionName string, un *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return un, nil
}
