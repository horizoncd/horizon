package getter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Helper struct {
	inner workload.Workload
}

func New(inner workload.Workload) *Helper {
	return &Helper{inner: inner}
}

func (w *Helper) GetSteps(node *v1alpha1.ResourceNode, client *kube.Client) (*workload.Step, error) {
	releaser, ok := w.inner.(workload.GreyscaleReleaser)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support greyscale release", reflect.TypeOf(w.inner))
	}
	steps, err := releaser.GetSteps(node, client)
	if err != nil {
		return nil, herrors.NewErrGetFailed(herrors.ResourceInK8S,
			fmt.Sprintf("failed to get steps: resource name = %v, err = %v", node.Name, err))
	}
	return steps, nil
}

// TODO: remove this after using informer
func (w *Helper) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	lister, ok := w.inner.(workload.PodsLister)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support list release", reflect.TypeOf(w.inner))
	}
	pods, err := lister.ListPods(node, client)
	if err != nil {
		return nil,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to list pods: resource name = %v, err = %v", node.Name, err))
	}
	return pods, nil
}

func (w *Helper) IsHealthy(node *v1alpha1.ResourceNode, client *kube.Client) (bool, error) {
	statusGetter, ok := w.inner.(workload.HealthStatusGetter)
	if !ok {
		return true, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support get health status", reflect.TypeOf(w.inner))
	}

	isHealthy, err := statusGetter.IsHealthy(node, client)
	if err != nil {
		return isHealthy,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to get healthy: resource name = %v, err = %v", node.Name, err))
	}
	return isHealthy, nil
}

func (w *Helper) RunAction(actionName string, resourceName string,
	ns string, gvr schema.GroupVersionResource, client *kube.Client) error {
	un, err := client.Dynamic.Resource(gvr).
		Namespace(ns).Get(context.TODO(), resourceName, v1.GetOptions{})
	if err != nil {
		return herrors.NewErrGetFailed(herrors.ResourceInK8S,
			fmt.Sprintf("failed to get healthy: resource name = %v, err = %v", resourceName, err))
	}
	un, err = w.inner.Action(actionName, un)
	if err != nil {
		return err
	}

	_, err = client.Dynamic.Resource(gvr).
		Namespace(ns).Update(context.TODO(), un, v1.UpdateOptions{})
	return err
}
