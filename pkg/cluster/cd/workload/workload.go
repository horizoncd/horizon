package workload

import (
	"time"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GreyscaleReleaser interface {
	Releaser

	GetSteps(un *unstructured.Unstructured,
		resourceTree map[string]*v1alpha1.ResourceNode, client *kube.Client) (*Step, error)
}

type Releaser interface {
	GetRevisions(un *unstructured.Unstructured,
		resourceTree map[string]*v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*Revision, error)
}

type PodsLister interface {
	ListPods(un *unstructured.Unstructured,
		resourceTree map[string]*v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error)
}

type HealthStatusGetter interface {
	IsHealthy(un *unstructured.Unstructured, resourceTree map[string]*v1alpha1.ResourceNode,
		client *kube.Client) (bool, error)
}

type DefaultHealthStatusGetter struct{}

func (s *DefaultHealthStatusGetter) IsHealthy(un *unstructured.Unstructured,
	resourceTree map[string]*v1alpha1.ResourceNode, client *kube.Client,
	image string, restartTime *time.Time) (bool, error) {
	return true, nil
}

type ActionProvider interface {
	Discovery()
	Actions()
	PerformActions()
}
