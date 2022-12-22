package getter

import (
	"reflect"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
)

type Getter struct {
	inner workload.Workload
}

func New(inner workload.Workload) *Getter {
	return &Getter{inner: inner}
}

func (w *Getter) GetSteps(node *v1alpha1.ResourceNode, client *kube.Client) (*workload.Step, error) {
	releaser, ok := w.inner.(workload.GreyscaleReleaser)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support greyscale release", reflect.TypeOf(w.inner))
	}
	steps, err := releaser.GetSteps(node, client)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to get steps"),
			"failed to get steps: resource name = %v, err = %v", node.Name, err)
	}
	return steps, nil
}

// TODO: remove this after using informer
func (w *Getter) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	lister, ok := w.inner.(workload.PodsLister)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support list release", reflect.TypeOf(w.inner))
	}
	pods, err := lister.ListPods(node, client)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to list pods"),
			"failed to list pods: resource name = %v, err = %v", node.Name, err)
	}
	return pods, nil
}

func (w *Getter) IsHealthy(node *v1alpha1.ResourceNode, client *kube.Client) (bool, error) {
	statusGetter, ok := w.inner.(workload.HealthStatusGetter)
	if !ok {
		return true, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support get health status", reflect.TypeOf(w.inner))
	}

	return statusGetter.IsHealthy(node, client)
}
