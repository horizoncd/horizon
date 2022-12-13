package ability

import (
	"reflect"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/cd/workload"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Workload struct {
	inner interface{}
}

func New(inner interface{}) *Workload {
	return &Workload{inner: inner}
}

func (w *Workload) GetSteps(un *unstructured.Unstructured, client *kube.Client) (*workload.Step, error) {
	releaser, ok := w.inner.(workload.GreyscaleReleaser)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support greyscale release", reflect.TypeOf(w.inner))
	}
	steps, err := releaser.GetSteps(un, client)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to get steps"),
			"failed to get steps: resource name = %v, err = %v", un.GetName(), err)
	}
	return steps, nil
}

// func (w *Workload) GetRevisionsOrListPods(un *unstructured.Unstructured,
// 	resourceTree []v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*workload.Revision, error) {
// 	releaser, ok := w.inner.(workload.Releaser)

// 	if !ok {
// 		pods, err := w.ListPods(un, resourceTree, client)
// 		if err != nil {
// 			return "", nil, err
// 		}
// 		return "current", map[string]*workload.Revision{"current": {Pods: pods}}, nil
// 	}
// 	return releaser.GetRevisions(un, resourceTree, client)
// }

// TODO: remove this after using informer
func (w *Workload) ListPods(un *unstructured.Unstructured,
	resourceTree []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	lister, ok := w.inner.(workload.PodsLister)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support list release", reflect.TypeOf(w.inner))
	}
	pods, err := lister.ListPods(un, resourceTree, client)
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S, "failed to list pods"),
			"failed to list pods: resource name = %v, err = %v", un.GetName(), err)
	}
	return pods, nil
}

func (w *Workload) IsHealthy(un *unstructured.Unstructured, client *kube.Client) (bool, error) {
	statusGetter, ok := w.inner.(workload.HealthStatusGetter)
	if !ok {
		return true, nil
	}
	return statusGetter.IsHealthy(un, client)
}
