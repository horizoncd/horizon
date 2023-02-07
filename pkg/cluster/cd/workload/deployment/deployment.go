package deployment

import (
	"context"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

func init() {
	workload.Register(ability)
}

// please refer to github.com/horizoncd/horizon/pkg/cluster/cd/workload/workload.go
var ability = &deployment{}

type deployment struct{}

func (*deployment) MatchGK(gk string) bool {
	return "apps/Deployment" == gk
}

func (*deployment) getDeployByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*v1.Deployment, error) {
	deploy, err := client.Basic.AppsV1().Deployments(node.Namespace).Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get deployment in k8s"),
			"failed to get deployment in k8s: deployment = %s, ns = %v, err = %v", node.Name, node.Namespace, err)
	}
	return deploy, nil
}

func (d *deployment) IsHealthy(node *v1alpha1.ResourceNode,
	client *kube.Client) (bool, error) {
	instance, err := d.getDeployByNode(node, client)
	if err != nil {
		return true, err
	}

	if instance.Status.ObservedGeneration != instance.Generation {
		return false, nil
	}

	return instance.Status.AvailableReplicas == *instance.Spec.Replicas, nil
}

func (d *deployment) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	instance, err := d.getDeployByNode(node, client)
	if err != nil {
		return nil, err
	}

	selector := instance.Spec.Selector
	pods, err := client.Basic.CoreV1().Pods(instance.ObjectMeta.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector:   polymorphichelpers.MakeLabels(selector.MatchLabels),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}
