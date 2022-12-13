package generic

import (
	"context"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var Ability = &Generic{}

type Generic struct {
}

func (g *Generic) ListPods(un *unstructured.Unstructured,
	resourceTree []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	var pods []corev1.Pod
	for _, node := range resourceTree {
		if node.Kind == "Pod" {
			pod, err := client.Basic.CoreV1().Pods(node.Namespace).Get(context.TODO(), node.Name, metav1.GetOptions{})
			if err != nil {
				return nil, perror.Wrapf(
					herrors.NewErrGetFailed(herrors.PodsInK8S, "failed to get Pods"),
					"failed to get pods: ns = %v, pod's name = %v, err = %v", node.Namespace, node.Name, err)
			}
			pods = append(pods, *pod)
		}
	}
	return pods, nil
}
