package workload

import (
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GreyscaleReleaser interface {
	Releaser

	GetSteps(un *unstructured.Unstructured, client *kube.Client) (*Step, error)
}

type Releaser interface {
	GetRevisions(un *unstructured.Unstructured,
		resourceTree []v1alpha1.ResourceNode, client *kube.Client) (string, map[string]*Revision, error)
}

type PodsLister interface {
	ListPods(un *unstructured.Unstructured,
		resourceTree []v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error)
}

type HealthStatusGetter interface {
	IsHealthy(un *unstructured.Unstructured, client *kube.Client) (bool, error)
}

type ActionProvider interface {
	Discovery()
	Actions()
	PerformActions()
}
